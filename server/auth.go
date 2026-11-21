package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return "rd_" + hex.EncodeToString(b), nil
}

func haskToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type authStore struct {
	db *pgxpool.Pool

	mu       sync.Mutex
	lastseen map[int64]time.Time
}

func newAuthStore(db *pgxpool.Pool) *authStore {
	return &authStore{db: db, lastseen: map[int64]time.Time{}}
}

func (a *authStore) create(ctx context.Context, name string) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}

	_, err = a.db.Exec(ctx,
		`INSERT INTO device_tokens (name, token_hash) VALUES ($1, $2)`,
		name, hashToken(token))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *authStore) revoke(ctx context.Context, name string) error {
	var id int64
	err := a.db.QueryRow(ctx,
		`UPDATE device_tokens SET revoked_at = now()
		WHERE name = $1 AND revoked_at IS NULL
		RETURNING id`, name).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("no active token named %q", name)
	}
	return err
}

func (a *authStore) list(ctx context.Context) error {
	rows, err := a.db.Query(ctx, `
		SELECT id, name, created_at, last_used_at, revoked_at
		FROM device_tokens ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Printf("%-4s %-12s %-17s %-17s %s\n", "ID", "NAME", "CREATED", "LAST USED", "STATUS")
	for rows.Next() {
		var (
			id                int64
			name              string
			created           time.Time
			lastUsed, revoked *time.Time
		)
		if err := rows.Scan(&id, &name, &created, &lastUsed, &revoked); err != nil {
			return err
		}
		status := "active"
		if revoked != nil {
			status = "revoked " + revoked.Format("2006-01-02")
		}
		fmt.Printf("%-4d %-12s %-17s %-17s %s\n",
			id, name, created.Format("2006-01-02 15:04"), fmtTime(lastUsed), status)
	}
	return rows.Err()
}

func fmtTime(t *time.Time) string {
	if t == nil {
		return "never"
	}
	return t.Format("2006-01-02 15:04")
}

func (a *authStore) check(ctx context.Context, token string) (string, bool) {
	var (
		id   int64
		name string
	)

	err := a.db.QueryRow(ctx, `
		SELECT id, name FROM device_token
		WHERE token_hash = $1 AND revoked_at IS NULL`,
		haskToken(token)).Scan(&id, &name)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Printf("auth: lookup failed: %w", err)
		}
		return "", false
	}
	a.touch(id)
	return name, true
}

func (a *authStore) touch(id int64) {
	a.mu.Lock()
	recent := time.Since(a.lastseen[id]) < time.Minute
	if !recent {
		a.lastseen[id] = time.Now()
	}
	a.mu.Unlock()

	if recent {
		return
	}

	go func() {
		_, err := a.db.Exec(context.Background(),
			`UPDATE device_tokens SET last_used_at = now() WHERE id = $1`, id)
		if err != nil {
			log.Printf("auth: update last_used_at: %v", err)
		}
	}()
}

var publicPaths = map[string]bool{
	"/health": true,
}
