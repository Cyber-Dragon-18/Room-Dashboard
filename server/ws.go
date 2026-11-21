package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

type event struct {
	Type     string         `json:"type"`
	DeviceID int64          `json:"device_id,omitempty"`
	Name     string         `json:"name,omitempty"`
	Changes  map[string]any `json:"changes,omitempty"`
}

func (h *hub) publish(e event) {
	b, err := json.Marshal(e)
	if err != nil {
		log.Printf("publish: %v", err)
		return
	}

	h.broadcast <- b
}

//GET /ws?token=rd_....

func (s *server) handleWS(w http.ResponseWriter, r *http.Request) {

	name, ok := s.auth.check(r.Context(), r.URL.Query().Get("token"))
	if !ok {
		http.Error(w, "invalid or missing token", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("ws accepted: %v", err)
		return
	}
	defer conn.CloseNow()

	c := &client{send: make(chan []byte, 16)}
	s.hub.register <- c
	defer func() { s.hub.unregister <- c }()
	log.Printf("ws: %s connected", name)
	defer log.Printf("ws: %s disconnected", name)

	ctx := conn.CloseRead(r.Context())

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Write(wctx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				return
			}

		case <-ticker.C:
			pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Ping(pctx)
			cancel()
			if err != nil {
				return
			}

		case <-ctx.Done():
			return
		}
	}

}
