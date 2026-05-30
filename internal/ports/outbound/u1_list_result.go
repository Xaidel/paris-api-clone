package ports

import "time"

// U1ListResult exposes a U1 list entry to inbound adapters.
type U1ListResult struct {
	ID                    string    `json:"id"`
	Sector                string    `json:"sector"`
	EligibleOperationType string    `json:"eligible_operation_type"`
	ConditionGuidance     string    `json:"condition_guidance"`
	CreatedBy             string    `json:"created_by"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// ListU1ListResult returns all U1 list entries.
type ListU1ListResult struct {
	Entries []U1ListResult `json:"entries"`
}
