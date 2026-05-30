package domain

import "testing"

// TestDomainErrorError verifies the domain error error behavior and the expected outcome asserted below.
func TestDomainErrorError(t *testing.T) {
	t.Parallel()

	err := &DomainError{Code: "TEST", Message: "example message"}
	if got := err.Error(); got != "[TEST] example message" {
		t.Fatalf("DomainError.Error() = %q", got)
	}
}

// TestValidationErrorFields verifies the validation error fields behavior and the expected outcome asserted below.
func TestValidationErrorFields(t *testing.T) {
	t.Parallel()

	err := NewValidationError([]FieldValidationError{
		NewFieldValidationError("goods_description", "required", "goods_description is required"),
	})
	if err == nil {
		t.Fatal("NewValidationError() returned nil, want error")
	}

	if got := err.Error(); got != "validation failed" {
		t.Fatalf("ValidationError.Error() = %q, want %q", got, "validation failed")
	}

	fields := err.Fields()
	if len(fields) != 1 {
		t.Fatalf("len(err.Fields()) = %d, want %d", len(fields), 1)
	}

	if fields[0].Field() != "goods_description" {
		t.Fatalf("fields[0].Field() = %q, want %q", fields[0].Field(), "goods_description")
	}
}
