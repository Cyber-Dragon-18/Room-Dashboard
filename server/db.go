package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type deviceRow struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Driver     string `json:"driver"`
	ExternalID string `json:"-"`
}

func loadDevices(ctx context.Context, db *pgxpool.Pool) ([]deviceRow, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, type, driver, external_id
		FROM devices
		ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []deviceRow
	for rows.Next() {
		var d deviceRow
		if err := rows.Scan(&d.ID, &d.Name, &d.Type, &d.Driver, &d.ExternalID); err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}
