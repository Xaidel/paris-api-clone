package ports

// GroupResult exposes a group entry to inbound adapters.
type GroupResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ListGroupsResult returns all group entries.
type ListGroupsResult struct {
	Groups []GroupResult `json:"groups"`
}
