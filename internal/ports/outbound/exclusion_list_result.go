package ports

import "time"

// ExclusionListResult exposes a U2 exclusion list entry to inbound adapters.
type ExclusionListResult struct {
	ID           string    `json:"id"`
	ActivityType string    `json:"activity_type"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ListExclusionListResult returns all exclusion list entries.
type ListExclusionListResult struct {
	Entries []ExclusionListResult `json:"entries"`
}
