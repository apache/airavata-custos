package models

import "time"

type AuditEvent struct {
	ID        string    `json:"id" db:"id"`
	EventType string    `json:"event_type" db:"event_type"` // e.g., "COMPUTE_ALLOCATION_CREATED", "COMPUTE_ALLOCATION_UPDATED", "COMPUTE_ALLOCATION_DELETED", etc.
	EventTime time.Time `json:"event_time" db:"event_time"`
	EntityID  string    `json:"entity_id" db:"entity_id"` // The ID of the entity associated with the event, e.g., the compute allocation ID.
	Details   string    `json:"details" db:"details"`     // Additional details about the event, stored as a JSON string or plain text.
}
