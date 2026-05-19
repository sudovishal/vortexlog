-- +goose Up
CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    service_name TEXT NOT NULL,
    log_level TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    ingested_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

-- +goose Down
DROP TABLE logs;
