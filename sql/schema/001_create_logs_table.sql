-- +goose Up
CREATE TABLE logs (
    id BIGSERIAL PRIMARY KEY,
    service_name TEXT NOT NULL,
    log_level TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

-- +goose Down
DROP TABLE logs;
