-- +goose Up
-- One row per dashboard page ("home", ...). widgets = the list of placed widgets.
CREATE TABLE layouts (
    page       TEXT        PRIMARY KEY,
    widgets    JSONB       NOT NULL DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE layouts;
