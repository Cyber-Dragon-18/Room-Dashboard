package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/jackc/pgx/v5"
)

// placedWidget mirrors PlacedWidget in the web app (widgets/types.ts).
type placedWidget struct {
	Key      string         `json:"key"`
	Widget   string         `json:"widget"`
	X        int            `json:"x"`
	Y        int            `json:"y"`
	W        int            `json:"w"`
	H        int            `json:"h"`
	Settings map[string]any `json:"settings,omitempty"`
}

const (
	gridColumns = 12
	maxRows     = 50
)

// Page names become part of the URL and the database: keep them simple.
var validPage = regexp.MustCompile(`^[a-z0-9-]{1,32}$`)

// GET /api/layouts/{page}
func (s *server) handleGetLayout(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	if !validPage.MatchString(page) {
		http.Error(w, "invalid page name", http.StatusBadRequest)
		return
	}

	var raw []byte
	err := s.db.QueryRow(r.Context(),
		`SELECT widgets FROM layouts WHERE page = $1`, page).Scan(&raw)
	if errors.Is(err, pgx.ErrNoRows) {
		// Never saved: the web app falls back to its default layout
		http.Error(w, "no saved layout", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "could not load layout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var widgets []placedWidget
	if err := json.Unmarshal(raw, &widgets); err != nil {
		http.Error(w, "stored layout is corrupt: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"page": page, "widgets": widgets})
}

// PUT /api/layouts/{page}   {"widgets": [...]}
func (s *server) handleSaveLayout(w http.ResponseWriter, r *http.Request) {
	page := r.PathValue("page")
	if !validPage.MatchString(page) {
		http.Error(w, "invalid page name", http.StatusBadRequest)
		return
	}

	var req struct {
		Widgets []placedWidget `json:"widgets"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Widgets == nil {
		http.Error(w, `body must be {"widgets": [...]}`, http.StatusBadRequest)
		return
	}
	if err := validateLayout(req.Widgets); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(req.Widgets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// "Upsert": insert, or update if this page already has a row
	_, err = s.db.Exec(r.Context(), `
		INSERT INTO layouts (page, widgets, updated_at) VALUES ($1, $2, now())
		ON CONFLICT (page) DO UPDATE SET widgets = EXCLUDED.widgets, updated_at = now()`,
		page, data)
	if err != nil {
		http.Error(w, "could not save layout: "+err.Error(), http.StatusInternalServerError)
		return
	}

	s.hub.publish(event{Type: "layout.updated", Name: page})
	w.WriteHeader(http.StatusNoContent)
}

// validateLayout rejects layouts that would break the grid.
func validateLayout(widgets []placedWidget) error {
	seen := map[string]bool{}
	for i, wd := range widgets {
		switch {
		case wd.Key == "" || wd.Widget == "":
			return fmt.Errorf("widget %d: key and widget are required", i)
		case seen[wd.Key]:
			return fmt.Errorf("widget %d: duplicate key %q", i, wd.Key)
		case wd.W < 1 || wd.H < 1:
			return fmt.Errorf("widget %q: w and h must be at least 1", wd.Key)
		case wd.X < 1 || wd.X+wd.W-1 > gridColumns:
			return fmt.Errorf("widget %q: doesn't fit in %d columns", wd.Key, gridColumns)
		case wd.Y < 1 || wd.Y+wd.H-1 > maxRows:
			return fmt.Errorf("widget %q: rows must be between 1 and %d", wd.Key, maxRows)
		}
		seen[wd.Key] = true
	}
	return nil
}
