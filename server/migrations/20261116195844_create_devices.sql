-- +goose Up
CREATE TABLE devices (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('light', 'socket')),
    driver TEXT NOT NULL,
    external_id TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (driver, external_id)
);

-- +goose Down
DROP TABLE devices;