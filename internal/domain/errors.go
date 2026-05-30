package domain

import (
	"errors"
	"fmt"
	"slices"
)

// DomainError represents a domain invariant error.
type DomainError struct {
	Code    string
	Message string
}

// FieldValidationError describes one invalid input field.
type FieldValidationError struct {
	field   string
	code    string
	message string
}

// ValidationError describes one or more field validation failures.
type ValidationError struct {
	fields []FieldValidationError
}

// Error returns the domain error string.
func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewFieldValidationError builds a field validation error value.
func NewFieldValidationError(field, code, message string) FieldValidationError {
	return FieldValidationError{field: field, code: code, message: message}
}

// Field returns the invalid field name.
func (e FieldValidationError) Field() string {
	return e.field
}

// Code returns the machine-readable validation code.
func (e FieldValidationError) Code() string {
	return e.code
}

// Message returns the validation failure message.
func (e FieldValidationError) Message() string {
	return e.message
}

// NewValidationError builds a validation error when any field failures exist.
func NewValidationError(fields []FieldValidationError) *ValidationError {
	if len(fields) == 0 {
		return nil
	}

	return &ValidationError{fields: slices.Clone(fields)}
}

// Error returns the validation error string.
func (e *ValidationError) Error() string {
	return "validation failed"
}

// Fields returns the validation field failures.
func (e *ValidationError) Fields() []FieldValidationError {
	if e == nil {
		return nil
	}

	return slices.Clone(e.fields)
}

var (
	ErrInvalidUserID                    = &DomainError{Code: "INVALID_USER_ID", Message: "user id is invalid"}
	ErrInvalidBugReportID               = &DomainError{Code: "INVALID_BUG_REPORT_ID", Message: "bug report id is invalid"}
	ErrInvalidFeedbackID                = &DomainError{Code: "INVALID_FEEDBACK_ID", Message: "feedback id is invalid"}
	ErrInvalidFeedbackKind              = &DomainError{Code: "INVALID_FEEDBACK_KIND", Message: "feedback kind must be one of: thumbs_up, thumbs_down"}
	ErrInvalidEventID                   = &DomainError{Code: "INVALID_EVENT_ID", Message: "event id is invalid"}
	ErrInvalidUploadID                  = &DomainError{Code: "INVALID_UPLOAD_ID", Message: "upload id is invalid"}
	ErrInvalidTransactionID             = &DomainError{Code: "INVALID_TRANSACTION_ID", Message: "transaction id is invalid"}
	ErrInvalidExclusionListID           = &DomainError{Code: "INVALID_EXCLUSION_LIST_ID", Message: "exclusion list id is invalid"}
	ErrInvalidU1ListID                  = &DomainError{Code: "INVALID_U1_LIST_ID", Message: "u1 list id is invalid"}
	ErrInvalidSectorID                  = &DomainError{Code: "INVALID_SECTOR_ID", Message: "sector id is invalid"}
	ErrInvalidGroupID                   = &DomainError{Code: "INVALID_GROUP_ID", Message: "group id is invalid"}
	ErrInvalidUsername                  = &DomainError{Code: "INVALID_USERNAME", Message: "username is required"}
	ErrInvalidFirstName                 = &DomainError{Code: "INVALID_FIRST_NAME", Message: "first name is required"}
	ErrInvalidLastName                  = &DomainError{Code: "INVALID_LAST_NAME", Message: "last name is required"}
	ErrInvalidGroupName                 = &DomainError{Code: "INVALID_GROUP_NAME", Message: "group name is required"}
	ErrInvalidActivityType              = &DomainError{Code: "INVALID_ACTIVITY_TYPE", Message: "activity type is required"}
	ErrInvalidSector                    = &DomainError{Code: "INVALID_SECTOR", Message: "sector is required"}
	ErrInvalidSectorType                = &DomainError{Code: "INVALID_SECTOR_TYPE", Message: "sector type must be one of: PA Aligned, High Emitting"}
	ErrInvalidSectorName                = &DomainError{Code: "INVALID_SECTOR_NAME", Message: "sector name is required"}
	ErrInvalidSectorDescription         = &DomainError{Code: "INVALID_SECTOR_DESCRIPTION", Message: "sector description is required"}
	ErrInvalidEligibleOperationType     = &DomainError{Code: "INVALID_ELIGIBLE_OPERATION_TYPE", Message: "eligible operation type is required"}
	ErrInvalidBugReportTitle            = &DomainError{Code: "INVALID_BUG_REPORT_TITLE", Message: "bug report title is required"}
	ErrInvalidBugReportDescription      = &DomainError{Code: "INVALID_BUG_REPORT_DESCRIPTION", Message: "bug report description is required"}
	ErrInvalidBugReportStatus           = &DomainError{Code: "INVALID_BUG_REPORT_STATUS", Message: "bug report status must be one of: Open, Closed"}
	ErrInvalidConditionGuidance         = &DomainError{Code: "INVALID_CONDITION_GUIDANCE", Message: "condition guidance is required"}
	ErrInvalidPassword                  = &DomainError{Code: "INVALID_PASSWORD", Message: "password must be at least 8 characters"}
	ErrInvalidPasswordHash              = &DomainError{Code: "INVALID_PASSWORD_HASH", Message: "password hash is required"}
	ErrInvalidActorUserID               = &DomainError{Code: "INVALID_ACTOR_USER_ID", Message: "actor user id is required"}
	ErrInvalidActorGroupID              = &DomainError{Code: "INVALID_ACTOR_GROUP_ID", Message: "actor group id is required"}
	ErrUnknownActorUserID               = &DomainError{Code: "UNKNOWN_ACTOR_USER_ID", Message: "actor user id was not found"}
	ErrUnknownActorGroupID              = &DomainError{Code: "UNKNOWN_ACTOR_GROUP_ID", Message: "actor group id was not found"}
	ErrInvalidEventType                 = &DomainError{Code: "INVALID_EVENT_TYPE", Message: "event type is required"}
	ErrInvalidEventData                 = &DomainError{Code: "INVALID_EVENT_DATA", Message: "event data must be valid json"}
	ErrInvalidTimestamp                 = &DomainError{Code: "INVALID_TIMESTAMP", Message: "timestamp is required"}
	ErrInvalidUploadFile                = &DomainError{Code: "INVALID_UPLOAD_FILE", Message: "upload file name is required"}
	ErrInvalidFileFormat                = &DomainError{Code: "INVALID_FILE_FORMAT", Message: "file format is invalid"}
	ErrInvalidFileHash                  = &DomainError{Code: "INVALID_FILE_HASH", Message: "file hash is invalid"}
	ErrInvalidStorage                   = &DomainError{Code: "INVALID_STORAGE", Message: "storage reference is invalid"}
	ErrInvalidSchema                    = &DomainError{Code: "INVALID_SCHEMA", Message: "schema version is required"}
	ErrInvalidColumnIndex               = &DomainError{Code: "INVALID_COLUMN_INDEX", Message: "column index must be greater than or equal to 0"}
	ErrInvalidRowCount                  = &DomainError{Code: "INVALID_ROW_COUNT", Message: "row count must be zero for failed uploads or greater than zero for uploaded uploads"}
	ErrInvalidTransaction               = &DomainError{Code: "INVALID_TRANSACTION", Message: "transaction row is invalid"}
	ErrInvalidTransactionClassification = &DomainError{Code: "INVALID_TRANSACTION_CLASSIFICATION", Message: "transaction classification must be one of: unclassified, aligned, not-aligned, next_step"}
	ErrInvalidTransactionStatus         = &DomainError{Code: "INVALID_TRANSACTION_STATUS", Message: "transaction status must be one of: pending, processing, ai-reviewed, professionally-reviewed, failed"}
	ErrInvalidTransactionUploadStatus   = &DomainError{Code: "INVALID_TRANSACTION_UPLOAD_STATUS", Message: "transaction upload status must be one of: uploaded, failed"}
	ErrDuplicateUpload                  = &DomainError{Code: "DUPLICATE_UPLOAD", Message: "upload file has already been ingested"}
	ErrEmptyGoodsDescription            = errors.New("keyword scoring: goods description is empty")
	ErrEmptyListEntries                 = errors.New("keyword scoring: list entries are empty")
)
