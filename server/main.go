package main

// main.go — starts the room dashboard server.
//
// Needs in .env (or the environment):
//   DATABASE_URL=postgres://room:room@localhost:5432/room?sslmode=disable
//   TUYA_ID=...  TUYA_SECRET=...   (only if you have tuya devices)
//   PORT=8080                      (optional)
//
// Token commands (don't start the server):
//   go run . token create <name> | token list | token revoke <name>

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Cyber-Dragon-18/room-dashboard/server/devices"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	// 1. Connect to Postgres
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("database config: ", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal("database unreachable (is the room-db container running?): ", err)
	}

	auth := newAuthStore(db)

	// 2. "go run . token ..." → run the token command and exit (no server)
	if len(os.Args) > 1 && os.Args[1] == "token" {
		if err := auth.runCommand(ctx, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	// 3. Load the device rows
	rows, err := loadDevices(ctx, db)
	if err != nil {
		log.Fatal("load devices: ", err)
	}

	// 4. Build a driver for every light, using the registry
	h := newHub()
	go h.run() // start the hub's loop

	s := &server{
		devices: rows,
		byID:    map[int64]deviceRow{},
		lights:  map[int64]devices.Light{},
		hub:     h,
		auth:    auth,
		db:      db,
	}
	for _, row := range rows {
		s.byID[row.ID] = row
		if row.Type != "light" {
			continue // sockets come later
		}
		light, err := devices.NewLight(row.Driver, row.ExternalID)
		if err != nil {
			log.Fatalf("device %d (%s): %v", row.ID, row.Name, err)
		}
		s.lights[row.ID] = light
		log.Printf("loaded light %d: %s [%s]", row.ID, row.Name, row.Driver)
	}

	// 5. Load the scenes
	scenes, err := loadScenes(ctx, db)
	if err != nil {
		log.Fatal("load scenes: ", err)
	}
	s.scenes = map[string]scene{}
	for _, sc := range scenes {
		s.scenes[sc.Name] = sc
	}
	log.Printf("loaded %d scenes", len(s.scenes))

	// 6. Routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/devices", s.handleListDevices)
	mux.HandleFunc("GET /api/devices/{id}", s.handleGetDevice)
	mux.HandleFunc("POST /api/devices/{id}/power", s.handleSetPower)
	mux.HandleFunc("POST /api/devices/{id}/brightness", s.handleSetBrightness)
	mux.HandleFunc("POST /api/devices/{id}/color", s.handleSetColor)
	mux.HandleFunc("POST /api/devices/{id}/white", s.handleSetWhite)
	mux.HandleFunc("GET /api/scenes", s.handleListScenes)
	mux.HandleFunc("POST /api/scenes/{name}/run", s.handleRunScene)
	mux.HandleFunc("GET /api/layouts/{page}", s.handleGetLayout)
	mux.HandleFunc("PUT /api/layouts/{page}", s.handleSaveLayout)
	mux.HandleFunc("GET /ws", s.handleWS)

	// 7. Start — the whole mux is wrapped by the auth middleware,
	// so every route needs a token except the public paths (/health).
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, auth.middleware(mux)))
}
