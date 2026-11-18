package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Cyber-Dragon-18/room-dashboard/server/devices"
)

type server struct {
	devices []deviceRow
	byID    map[int64]deviceRow
	lights  map[int64]devices.Light
}

// GET /api/devices
func (s *server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.devices)
}

// GET /api/devices/{id}
func (s *server) handleGetDevices(w http.ResponseWriter, r *http.Request) {
	row, light, ok := s.lightFromPath(w, r)
	if !ok {
		return
	}

	st, err := light.State()
	if err != nil {
		http.Error(w, "could not read device state: "+err.Error(), http.StatusBadGateway)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"device": row,
		"state":  st,
	})
}

// POST /api/devices/{id}/power
func (s *server) handleSetPower(w http.ResponseWriter, r *http.Request) {
	_, light, ok := s.lightFromPath(w, r)
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

	w.WriteHeader(http.StatusNoContent)
}

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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON; %v", err)
	}
}
