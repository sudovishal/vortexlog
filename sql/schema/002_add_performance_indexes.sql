-- +goose Up
-- Enable trigram extension before we can use it for text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;


-- Composite B-TREE INDEX
CREATE INDEX idx_log_service_level_time
ON logs (service_name, log_level, ingested_at);

-- GIN INDEX(JSONB optimization)
-- The jsonb_path_ops operator makes searching for specific key-value pairs 
-- inside your JSONB blob incredibly fast and keeps the index size smaller
CREATE INDEX idx_logs_metadata_gin
ON logs USING gin(metadata jsonb_path_ops);

-- TRIGRAM index(WILD CARD SEARCH)
CREATE INDEX idx_logs_message_trgm
ON logs USING GIN (message gin_trgm_ops);



-- +goose Down
DROP INDEX IF EXISTS idx_logs_message_trgm;
DROP INDEX IF EXISTS idx_logs_metadata_gin;
DROP INDEX IF EXISTS idx_log_service_level_time;
DROP EXTENSION IF EXISTS pg_trgm;
