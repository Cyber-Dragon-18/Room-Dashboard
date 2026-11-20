-- +goose Up
CREATE TABLE scenes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    actions JSONB NOT NULL
);

-- +goose Down
DROP TABLE scenes;
