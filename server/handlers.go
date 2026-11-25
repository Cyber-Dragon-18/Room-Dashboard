package main

// handlers.go — the HTTP API for devices.
//
//	GET  /api/devices                  list all devices
//	GET  /api/devices/{id}             one device + its current state
//	POST /api/devices/{id}/power       {"on": true}
//	POST /api/devices/{id}/brightness  {"pct": 50}
//	POST /api/devices/{id}/color       {"r": 0, "g": 0, "b": 255}   (colour mode, RGB LEDs)
//	POST /api/devices/{id}/white       {"temp": 70}                  (white mode, 0 warm - 100 cool)
//
// Every command handler follows the same order:
// read + validate → do the action → if it worked, publish the event → respond.

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Cyber-Dragon-18/room-dashboard/server/devices"
	"github.com/jackc/pgx/v5/pgxpool"
)

// server holds everything the handlers need.
type server struct {
	devices []deviceRow             // all rows, for listing
	byID    map[int64]deviceRow     // quick lookup by id
	lights  map[int64]devices.Light // the driver object for each light
	scenes  map[string]scene        // scenes by name
	hub     *hub                    // live updates (WebSocket)
	auth    *authStore              // token checks
	db      *pgxpool.Pool           // for things read on demand (layouts)
}

// GET /api/devices
func (s *server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.devices)
}

// GET /api/devices/{id}
func (s *server) handleGetDevice(w http.ResponseWriter, r *http.Request) {
	row, light, ok := s.lightFromPath(w, r)
	if !ok {
		return
	}
	st, err := light.State()
	if err != nil {
		http.Error(w, "could not read device state: "+err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"device": row, "state": st})
}

// POST /api/devices/{id}/power   {"on": true}
func (s *server) handleSetPower(w http.ResponseWriter, r *http.Request) {
	row, light, ok := s.lightFromPath(w, r)
	if !ok {
		return
	}

	var req struct {
		On *bool `json:"on"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.On == nil {
		http.Error(w, `body must be {"on": true} or {"on": false}`, http.StatusBadRequest)
		return
	}

	if err := light.SetPower(*req.On); err != nil {
		http.Error(w, "could not set power: "+err.Error(), http.StatusBadGateway)
		return
	}

	s.hub.publish(event{
		Type:     "device.updated",
		DeviceID: row.ID,
		Changes:  map[string]any{"on": *req.On},
	})
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/devices/{id}/brightness   {"pct": 50}
func (s *server) handleSetBrightness(w http.ResponseWriter, r *http.Request) {
	row, light, ok := s.lightFromPath(w, r)
	if !ok {
		return
	}

	var req struct {
		Pct *int `json:"pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Pct == nil {
		http.Error(w, `body must be {"pct": 0-100}`, http.StatusBadRequest)
		return
	}
	if *req.Pct < 0 || *req.Pct > 100 {
		http.Error(w, "pct must be between 0 and 100", http.StatusBadRequest)
		return
	}

	if err := light.SetBrightness(*req.Pct); err != nil {
		http.Error(w, "could not set brightness: "+err.Error(), http.StatusBadGateway)
		return
	}

	s.hub.publish(event{
		Type:     "device.updated",
		DeviceID: row.ID,
		Changes:  map[string]any{"brightness": *req.Pct},
	})
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/devices/{id}/color   {"r": 0, "g": 0, "b": 255}
func (s *server) handleSetColor(w http.ResponseWriter, r *http.Request) {
	row, light, ok := s.lightFromPath(w, r)
	if !ok {
		return
	}

	var req struct {
		R *int `json:"r"`
		G *int `json:"g"`
		B *int `json:"b"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil ||
		req.R == nil || req.G == nil || req.B == nil {
		http.Error(w, `body must be {"r": 0-255, "g": 0-255, "b": 0-255}`, http.StatusBadRequest)
		return
	}
	// Check the range before converting: uint8(300) would silently become 44
	if !in255(*req.R) || !in255(*req.G) || !in255(*req.B) {
		http.Error(w, "r, g and b must be between 0 and 255", http.StatusBadRequest)
		return
	}

	if err := light.SetColor(uint8(*req.R), uint8(*req.G), uint8(*req.B)); err != nil {
		http.Error(w, "could not set color: "+err.Error(), http.StatusBadGateway)
		return
	}

	s.hub.publish(event{
		Type:     "device.updated",
		DeviceID: row.ID,
		Changes: map[string]any{
			"mode":  "colour", // the lamp switched to its RGB LEDs
			"color": map[string]int{"r": *req.R, "g": *req.G, "b": *req.B},
		},
	})
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/devices/{id}/white   {"temp": 70}   0 = warm, 100 = cool
func (s *server) handleSetWhite(w http.ResponseWriter, r *http.Request) {
	row, light, ok := s.lightFromPath(w, r)
	if !ok {
		return
	}

	var req struct {
		Temp *int `json:"temp"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Temp == nil {
		http.Error(w, `body must be {"temp": 0-100}`, http.StatusBadRequest)
		return
	}
	if *req.Temp < 0 || *req.Temp > 100 {
		http.Error(w, "temp must be between 0 (warm) and 100 (cool)", http.StatusBadRequest)
		return
	}

	if err := light.SetWhite(*req.Temp); err != nil {
		http.Error(w, "could not set white: "+err.Error(), http.StatusBadGateway)
		return
	}

	s.hub.publish(event{
		Type:     "device.updated",
		DeviceID: row.ID,
		Changes:  map[string]any{"mode": "white", "temp": *req.Temp},
	})
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleListScenes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.scenes)
}

// runScene applies every action of a scene.
func (s *server) runScene(sc scene) []string {
	var failed []string

	for _, a := range sc.Actions {
		light, ok := s.lights[a.DeviceID]
		if !ok {
			failed = append(failed, fmt.Sprintf("device %d: not found", a.DeviceID))
			continue
		}

		if a.Power != nil {
			if err := light.SetPower(*a.Power); err != nil {
				failed = append(failed, fmt.Sprintf("device %d power: %v", a.DeviceID, err))
				continue
			}
		}

		switch {
		case a.Temp != nil:
			if err := light.SetWhite(*a.Temp); err != nil {
				failed = append(failed, fmt.Sprintf("device %d white: %v", a.DeviceID, err))
			}
		case a.Color != nil:
			if err := light.SetColor(a.Color.R, a.Color.G, a.Color.B); err != nil {
				failed = append(failed, fmt.Sprintf("device %d color: %v", a.DeviceID, err))
			}
		}

		if a.Brightness != nil {
			if err := light.SetBrightness(*a.Brightness); err != nil {
				failed = append(failed, fmt.Sprintf("device %d brightness: %v", a.DeviceID, err))
			}
		}
	}
	return failed
}

// POST /api/scenes/{name}/run
func (s *server) handleRunScene(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	sc, ok := s.scenes[name]
	if !ok {
		http.Error(w, "scene not found", http.StatusNotFound)
		return
	}

	if failed := s.runScene(sc); len(failed) > 0 {
		writeJSON(w, http.StatusBadGateway, map[string]any{"scene": name, "errors": failed})
		return
	}

	s.hub.publish(event{Type: "scene.ran", Name: name})
	w.WriteHeader(http.StatusNoContent)
}

func in255(v int) bool { return v >= 0 && v <= 255 }

// lightFromPath reads {id} from the URL and finds that light.
// If anything is wrong it writes the error response itself and returns ok=false.
func (s *server) lightFromPath(w http.ResponseWriter, r *http.Request) (deviceRow, devices.Light, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "device id must be a number", http.StatusBadRequest)
		return deviceRow{}, nil, false
	}
	row, found := s.byID[id]
	if !found {
		http.Error(w, "device not found", http.StatusNotFound)
		return deviceRow{}, nil, false
	}
	light, isLight := s.lights[id]
	if !isLight {
		http.Error(w, "device is not a light", http.StatusBadRequest)
		return deviceRow{}, nil, false
	}
	return row, light, true
}

// writeJSON sends v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}
