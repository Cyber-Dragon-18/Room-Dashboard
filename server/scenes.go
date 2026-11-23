package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type sceneAction struct {
	DeviceID   int64 `json:"device_id"`
	Power      *bool `json:"power"`
	Brightness *int  `json:"brightness"`
	Color      *struct {
		R uint8 `json:"r"`
		G uint8 `json:"g"`
		B uint8 `json:"b"`
	} `json:"color"`
	Temp *int `json:"temp"`
}

type scene struct {
	ID      int64         `json:"id"`
	Name    string        `json:"name"`
	Actions []sceneAction `json:"actions"`
}

func loadScenes(ctx context.Context, db *pgxpool.Pool) ([]scene, error) {
	rows, err := db.Query(ctx, `SELECT id, name, actions FROM scenes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []scene
	for rows.Next() {
		var (
			sc  scene
			raw []byte
		)
		if err := rows.Scan(&sc.ID, &sc.Name, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &sc.Actions); err != nil {
			return nil, fmt.Errorf("scene %q: bad actions JSON: %w", sc.Name, err)
		}
		list = append(list, sc)
	}
	return list, rows.Err()
}
