package ports

import "time"

// UserResult exposes a user to inbound adapters.
type UserResult struct {
	ID         string
	Username   string
	FirstName  string
	MiddleName *string
	LastName   string
	GroupID    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
