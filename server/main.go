package main

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

	// Connecrt database
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("database config: ", err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal("database unreachable (is the room-db running?): ", err)
	}

	auth := newAuthStore(db)

	if len(os.Args) > 1 && os.Args[1] == "token" {
		if err := auth.runCommand(ctx, os.Args[2:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Load the device rows
	rows, err := loadDevices(ctx, db)
	if err != nil {
		log.Fatal("load devices: ", err)
	}

	// Build a driver for every light

	s := &server{
		devices: rows,
		byID:    map[int64]deviceRow{},
		lights:  map[int64]devices.Light{},
	}
	for _, row := range rows {
		s.byID[row.ID] = row
		if row.Type != "light" {
			continue
		}
		light, err := devices.NewLight(row.Driver, row.ExternalID)
		if err != nil {
			log.Fatalf("device %d (%s): %v", row.ID, row.Name, err)
		}
		s.lights[row.ID] = light
		log.Printf("loaded light %d: %s [%s]", row.ID, row.Name, row.Driver)
	}

	scenes, err := loadScenes(ctx, db)
	if err != nil {
		log.Fatal("load scenes: ", err)
	}
	s.scenes = map[string]scene{}
	for _, sc := range scenes {
		s.scenes[sc.Name] = sc
	}
	log.Printf("loaded %d scenes", len(s.scenes))

	h := newHub()
	go h.run()
	s.hub = h
	s.auth = auth

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/devices", s.handleListDevices)
	mux.HandleFunc("GET /api/devices/{id}", s.handleGetDevices)
	mux.HandleFunc("POST /api/devices/{id}/power", s.handleSetPower)
	mux.HandleFunc("POST /api/devices/{id}/bright", s.handleSetBrightness)
	mux.HandleFunc("POST /api/devices/{id}/color", s.handleSetColor)
	mux.HandleFunc("GET /api/scenes", s.handleListScenes)
	mux.HandleFunc("POST /api/scenes/{name}/run", s.handleRunScene)
	mux.HandleFunc("GET /ws", s.handleWS)

	//Start
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, auth.middleware(mux)))

}
