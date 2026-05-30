package adapters

import (
	"context"
	"regexp"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/pashagolub/pgxmock/v4"
)

// TestPostgresGroupRepository verifies the postgres group repository behavior and the expected outcome asserted below.
func TestPostgresGroupRepository(t *testing.T) {
	t.Parallel()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("pgxmock.NewPool() error = %v", err)
	}
	defer mock.Close()

	groupID, err := valueobjects.GroupIDFromString("01962b8f-aeb2-7e03-a8ff-1edce1300001")
	if err != nil {
		t.Fatalf("GroupIDFromString() error = %v", err)
	}

	group, err := entities.NewGroup(groupID, "superadmin")
	if err != nil {
		t.Fatalf("NewGroup() error = %v", err)
	}

	repository := NewPostgresGroupRepository(mock)
	mock.ExpectExec(regexp.QuoteMeta(createGroupQuery)).WithArgs(group.ID().String(), group.Name()).WillReturnResult(pgxmock.NewResult("INSERT", 1))
	if err := repository.Create(context.Background(), group); err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}
