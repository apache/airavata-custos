-- Add retry backoff support to processing events.

ALTER TABLE amie_processing_events
    ADD COLUMN next_retry_at TIMESTAMP(6) NULL DEFAULT NULL,
    ADD INDEX idx_amie_events_next_retry_at (next_retry_at);
