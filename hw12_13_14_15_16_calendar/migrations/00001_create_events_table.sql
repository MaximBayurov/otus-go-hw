-- +goose Up
-- +goose StatementBegin
CREATE TABLE events
(
    id          VARCHAR(36) PRIMARY KEY,
    title       TEXT,
    start_time  TIMESTAMP    NOT NULL,
    end_time    TIMESTAMP    NOT NULL,
    description TEXT,
    owner_id    VARCHAR(100) NOT NULL,
    notify_time TIMESTAMP    NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT valid_duration CHECK (end_time > start_time)
);

-- Индексы для оптимизации запросов
CREATE INDEX idx_events_date ON events (DATE(start_time));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events
-- +goose StatementEnd
