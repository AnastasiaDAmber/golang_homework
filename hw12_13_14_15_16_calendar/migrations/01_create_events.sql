-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title TEXT NOT NULL,
    at TIMESTAMPTZ NOT NULL,
    duration INTERVAL,
    description TEXT,
    user_id TEXT,
    notify_before INTERVAL
);

-- +goose Down
CREATE INDEX IF NOT EXISTS idx_events_at ON events (at);
CREATE INDEX IF NOT EXISTS idx_events_user_at ON events (user_id, at);
