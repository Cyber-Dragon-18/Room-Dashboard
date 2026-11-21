package main

import (
	"context"

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
}

type scene struct {
	ID      int64         `json:"id"`
	Name    string        `json:"string"`
	Actions []sceneAction `json:"actions"`
}

func loadScenes(ctx context.Context, db *pgxpool.Pool) ([]scene, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, actions
		FROM scenes
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []scene
	for rows.Next() {
		var s scene
		if err := rows.Scan(&s.ID, &s.Name, &s.Actions); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}
