package ports

import "time"

// SectorResult exposes a sector entry to inbound adapters.
type SectorResult struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListSectorsResult returns all sector entries.
type ListSectorsResult struct {
	Sectors []SectorResult `json:"sectors"`
}
