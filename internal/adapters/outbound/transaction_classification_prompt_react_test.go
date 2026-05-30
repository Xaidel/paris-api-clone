package adapters

import (
	"context"
	"strings"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/ports/outbound"
)

const testReactClassificationPromptTemplate = `## CHECK 1 — Not Aligned List
{% for entry in exclusionary_list %}
- {{ entry.activity_type }}
{% endfor %}
## CHECK 2 — Aligned List
| Sector | Eligible operation type | Conditions and guidance |
| :--- | :--- | :--- |
{% for entry in u1_list %}
| {{ entry.sector }}  | {{ entry.eligible_operation_type }}  | {{ entry.condition_guidance }}  |
{% endfor %}`

type reactPromptU1ListRepositoryStub struct {
	entries []*entities.U1ListEntry
	err     error
	filter  ports.U1ListFilter
}

func (s *reactPromptU1ListRepositoryStub) Create(context.Context, *entities.U1ListEntry, string) error {
	return nil
}

func (s *reactPromptU1ListRepositoryStub) FindByID(context.Context, valueobjects.U1ListID) (*entities.U1ListEntry, error) {
	return nil, nil
}

func (s *reactPromptU1ListRepositoryStub) List(_ context.Context, filter ports.U1ListFilter) ([]*entities.U1ListEntry, error) {
	s.filter = filter
	return s.entries, s.err
}

func (s *reactPromptU1ListRepositoryStub) Update(context.Context, *entities.U1ListEntry) error {
	return nil
}

func (s *reactPromptU1ListRepositoryStub) DeleteByID(context.Context, valueobjects.U1ListID) error {
	return nil
}

type reactPromptExclusionListRepositoryStub struct {
	entries []*entities.ExclusionListEntry
	err     error
}

func (s *reactPromptExclusionListRepositoryStub) Create(context.Context, *entities.ExclusionListEntry, string) error {
	return nil
}

func (s *reactPromptExclusionListRepositoryStub) FindByID(context.Context, valueobjects.ExclusionListID) (*entities.ExclusionListEntry, error) {
	return nil, nil
}

func (s *reactPromptExclusionListRepositoryStub) List(context.Context) ([]*entities.ExclusionListEntry, error) {
	return s.entries, s.err
}

func (s *reactPromptExclusionListRepositoryStub) Update(context.Context, *entities.ExclusionListEntry) error {
	return nil
}

func (s *reactPromptExclusionListRepositoryStub) DeleteByID(context.Context, valueobjects.ExclusionListID) error {
	return nil
}

// TestNewReActTransactionClassificationSystemPromptBuilderRendersListsFromRepositories verifies the new ReAct act transaction classification system prompt builder renders lists from repositories behavior and the expected outcome asserted below.
func TestNewReActTransactionClassificationSystemPromptBuilderRendersListsFromRepositories(t *testing.T) {
	t.Parallel()

	exclusionID, _ := valueobjects.ExclusionListIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	u1ID, _ := valueobjects.U1ListIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61")

	exclusionEntry, err := entities.NewExclusionListEntry(exclusionID, "Electricity generation from coal")
	if err != nil {
		t.Fatalf("NewExclusionListEntry() error = %v", err)
	}

	u1Entry, err := entities.NewU1ListEntry(
		u1ID,
		"Energy",
		"Generation of renewable energy (e.g., from wind, solar, wave power, etc.) with negligible lifecycle GHG emissions.",
		"Includes generation of heat or cooling.",
	)
	if err != nil {
		t.Fatalf("NewU1ListEntry() error = %v", err)
	}

	u1Repo := &reactPromptU1ListRepositoryStub{entries: []*entities.U1ListEntry{u1Entry}}
	exclusionRepo := &reactPromptExclusionListRepositoryStub{entries: []*entities.ExclusionListEntry{exclusionEntry}}
	builder := NewReActTransactionClassificationSystemPromptBuilder(testReactClassificationPromptTemplate, u1Repo, exclusionRepo)

	prompt, err := builder(context.Background())
	if err != nil {
		t.Fatalf("builder() error = %v", err)
	}

	if u1Repo.filter != (ports.U1ListFilter{}) {
		t.Fatalf("u1Repo.filter = %+v, want empty filter", u1Repo.filter)
	}

	if !strings.Contains(prompt, "- Electricity generation from coal") {
		t.Fatalf("prompt does not contain rendered exclusion list entry: %q", prompt)
	}

	if !strings.Contains(prompt, "| Energy  | Generation of renewable energy (e.g., from wind, solar, wave power, etc.) with negligible lifecycle GHG emissions.  | Includes generation of heat or cooling.  |") {
		t.Fatalf("prompt does not contain rendered U1 table row: %q", prompt)
	}

	if strings.Contains(prompt, "{% for entry in exclusionary_list %}") {
		t.Fatalf("prompt still contains exclusionary_list template syntax: %q", prompt)
	}

	if strings.Contains(prompt, "{% for entry in u1_list %}") {
		t.Fatalf("prompt still contains u1_list template syntax: %q", prompt)
	}
}
