package ports

import (
	"encoding/json"
	"time"
)

// AuditEventResult is the response DTO for a single audit event.
type AuditEventResult struct {
	ID         string          `json:"id"`
	Timestamp  time.Time       `json:"timestamp"`
	EventOwner string          `json:"event_owner"`
	EventType  string          `json:"event_type"`
	SessionID  string          `json:"session_id,omitempty"`
	UserID     string          `json:"user_id,omitempty"`
	GroupID    string          `json:"group_id,omitempty"`
	EventData  json.RawMessage `json:"event_data,omitempty"`
}

// ListAuditEventsResult is the response DTO for an audit event list.
type ListAuditEventsResult struct {
	Events []AuditEventResult `json:"events"`
}
