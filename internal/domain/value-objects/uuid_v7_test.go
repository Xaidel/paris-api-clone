package valueobjects

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/gyud-adb/paris-api/internal/domain"
)

type uuidV7IDTestCase struct {
	name    string
	newID   func() (string, error)
	parseID func(string) (string, error)
	wantErr error
}

func TestUUIDv7IDsRoundTrip(t *testing.T) {
	t.Parallel()

	for _, tc := range uuidV7IDTestCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			generated, err := tc.newID()
			if err != nil {
				t.Fatalf("newID() error = %v", err)
			}

			assertStringIsUUIDv7(t, generated)

			restored, err := tc.parseID(generated)
			if err != nil {
				t.Fatalf("parseID() error = %v", err)
			}

			if restored != generated {
				t.Fatalf("restored = %q, want %q", restored, generated)
			}
		})
	}
}

func TestUUIDv7IDsParseCanonicalLiteral(t *testing.T) {
	t.Parallel()

	const canonicalUUIDv7 = "01962b8f-aeb2-7e03-a8ff-1edce1300001"

	for _, tc := range uuidV7IDTestCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := tc.parseID(canonicalUUIDv7)
			if err != nil {
				t.Fatalf("parseID() error = %v", err)
			}

			if parsed != canonicalUUIDv7 {
				t.Fatalf("parsed = %q, want %q", parsed, canonicalUUIDv7)
			}
		})
	}
}

func TestUUIDv7IDsRejectInvalidInputs(t *testing.T) {
	t.Parallel()

	invalidInputs := []string{
		"",
		"abc",
		"bad-id",
		"not-a-uuid",
		"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
		"0123456789abcdef0123456789abcdef",
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	for _, tc := range uuidV7IDTestCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for _, input := range invalidInputs {
				input := input
				t.Run(input, func(t *testing.T) {
					t.Parallel()

					_, err := tc.parseID(input)
					if !errors.Is(err, tc.wantErr) {
						t.Fatalf("parseID(%q) error = %v, want %v", input, err, tc.wantErr)
					}
				})
			}
		})
	}
}

func TestUUIDv7IDsZeroValue(t *testing.T) {
	t.Parallel()

	if !(UserID{}).IsZero() {
		t.Fatal("expected zero-value UserID to report IsZero")
	}
	if !(UploadID{}).IsZero() {
		t.Fatal("expected zero-value UploadID to report IsZero")
	}
	if !(EventID{}).IsZero() {
		t.Fatal("expected zero-value EventID to report IsZero")
	}
	if !(TransactionID{}).IsZero() {
		t.Fatal("expected zero-value TransactionID to report IsZero")
	}
	if !(GroupID{}).IsZero() {
		t.Fatal("expected zero-value GroupID to report IsZero")
	}
	if !(SectorID{}).IsZero() {
		t.Fatal("expected zero-value SectorID to report IsZero")
	}
	if !(FeedbackID{}).IsZero() {
		t.Fatal("expected zero-value FeedbackID to report IsZero")
	}
	if !(ExclusionListID{}).IsZero() {
		t.Fatal("expected zero-value ExclusionListID to report IsZero")
	}
	if !(U1ListID{}).IsZero() {
		t.Fatal("expected zero-value U1ListID to report IsZero")
	}
	if !(BugReportID{}).IsZero() {
		t.Fatal("expected zero-value BugReportID to report IsZero")
	}
}

func TestUUIDv7IDsUseDistinctPhantomTypes(t *testing.T) {
	t.Parallel()

	type UserIDAlias = UserID
	type RawUserID = UUIDv7ID[userIDTag]
	type RawUploadID = UUIDv7ID[uploadIDTag]

	if reflect.TypeOf(UserID{}) == reflect.TypeOf(UploadID{}) {
		t.Fatal("expected UserID and UploadID to be distinct types")
	}

	if reflect.TypeOf(RawUserID{}) == reflect.TypeOf(RawUploadID{}) {
		t.Fatal("expected UUIDv7ID[userIDTag] and UUIDv7ID[uploadIDTag] to be distinct types")
	}

	if reflect.TypeOf(UserID{}) != reflect.TypeOf(UserIDAlias{}) {
		t.Fatal("expected a true alias of UserID to remain the same type")
	}

	if reflect.TypeOf(UserID{}) != reflect.TypeOf(RawUserID{}) {
		t.Fatal("expected UserID to be the same type as UUIDv7ID[userIDTag]")
	}

	if reflect.TypeOf(UploadID{}) != reflect.TypeOf(RawUploadID{}) {
		t.Fatal("expected UploadID to be the same type as UUIDv7ID[uploadIDTag]")
	}
}

func uuidV7IDTestCases() []uuidV7IDTestCase {
	return []uuidV7IDTestCase{
		{name: "UserID", newID: newIDToString(NewUserID), parseID: parseIDToString(UserIDFromString), wantErr: domain.ErrInvalidUserID},
		{name: "UploadID", newID: newIDToString(NewUploadID), parseID: parseIDToString(UploadIDFromString), wantErr: domain.ErrInvalidUploadID},
		{name: "EventID", newID: newIDToString(NewEventID), parseID: parseIDToString(EventIDFromString), wantErr: domain.ErrInvalidEventID},
		{name: "TransactionID", newID: newIDToString(NewTransactionID), parseID: parseIDToString(TransactionIDFromString), wantErr: domain.ErrInvalidTransactionID},
		{name: "GroupID", newID: newIDToString(NewGroupID), parseID: parseIDToString(GroupIDFromString), wantErr: domain.ErrInvalidGroupID},
		{name: "SectorID", newID: newIDToString(NewSectorID), parseID: parseIDToString(SectorIDFromString), wantErr: domain.ErrInvalidSectorID},
		{name: "FeedbackID", newID: newIDToString(NewFeedbackID), parseID: parseIDToString(FeedbackIDFromString), wantErr: domain.ErrInvalidFeedbackID},
		{name: "ExclusionListID", newID: newIDToString(NewExclusionListID), parseID: parseIDToString(ExclusionListIDFromString), wantErr: domain.ErrInvalidExclusionListID},
		{name: "U1ListID", newID: newIDToString(NewU1ListID), parseID: parseIDToString(U1ListIDFromString), wantErr: domain.ErrInvalidU1ListID},
		{name: "BugReportID", newID: newIDToString(NewBugReportID), parseID: parseIDToString(BugReportIDFromString), wantErr: domain.ErrInvalidBugReportID},
	}
}

func newIDToString[T any](newID func() (UUIDv7ID[T], error)) func() (string, error) {
	return func() (string, error) {
		id, err := newID()
		if err != nil {
			return "", err
		}

		return id.String(), nil
	}
}

func parseIDToString[T any](parseID func(string) (UUIDv7ID[T], error)) func(string) (string, error) {
	return func(value string) (string, error) {
		id, err := parseID(value)
		if err != nil {
			return "", err
		}

		return id.String(), nil
	}
}

func assertStringIsUUIDv7(t *testing.T, raw string) {
	t.Helper()

	parsedValue, err := uuid.Parse(raw)
	if err != nil {
		t.Fatalf("uuid.Parse(%q) error = %v", raw, err)
	}
	if parsedValue.String() != raw {
		t.Fatalf("parsedValue.String() = %q, want %q", parsedValue.String(), raw)
	}
	if parsedValue.Version() != uuid.Version(7) {
		t.Fatalf("parsedValue.Version() = %d, want %d", parsedValue.Version(), uuid.Version(7))
	}
	if parsedValue.Variant() != uuid.RFC4122 {
		t.Fatalf("parsedValue.Variant() = %v, want %v", parsedValue.Variant(), uuid.RFC4122)
	}
}
