package usecases

import "fmt"

// NotFoundError reports a missing application resource.
type NotFoundError struct {
	Resource string
	ID       string
}

// Error returns the not found error string.
func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s %q was not found", e.Resource, e.ID)
}

// ConflictError reports an application conflict.
type ConflictError struct {
	Resource string
	Reason   string
}

// Error returns the conflict error string.
func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s conflict: %s", e.Resource, e.Reason)
}

// ForbiddenError reports an application authorization failure.
type ForbiddenError struct {
	Resource string
	Reason   string
}

// Error returns the forbidden error string.
func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("%s forbidden: %s", e.Resource, e.Reason)
}
