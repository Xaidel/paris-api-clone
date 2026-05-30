package adapters

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

// NewReActTransactionClassificationSystemPromptBuilder builds a renderer for the
// ReAct classification system prompt template.
func NewReActTransactionClassificationSystemPromptBuilder(
	promptTemplate string,
	u1ListRepository ports.U1ListRepository,
	exclusionListRepository ports.ExclusionListRepository,
) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		if strings.TrimSpace(promptTemplate) == "" {
			return "", fmt.Errorf("react classification system prompt template is required")
		}

		if u1ListRepository == nil {
			return "", fmt.Errorf("u1 list repository is required")
		}

		if exclusionListRepository == nil {
			return "", fmt.Errorf("exclusion list repository is required")
		}

		exclusionEntries, err := exclusionListRepository.List(ctx)
		if err != nil {
			return "", fmt.Errorf("listing exclusion list entries for react prompt: %w", err)
		}

		u1Entries, err := u1ListRepository.List(ctx, ports.U1ListFilter{})
		if err != nil {
			return "", fmt.Errorf("listing u1 list entries for react prompt: %w", err)
		}

		tpl := prompt.FromMessages(schema.Jinja2, schema.SystemMessage(promptTemplate))
		renderedMessages, err := tpl.Format(ctx, map[string]any{
			"exclusionary_list": reactPromptExclusionaryList(exclusionEntries),
			"u1_list":           reactPromptU1List(u1Entries),
		})
		if err != nil {
			return "", fmt.Errorf("rendering react classification prompt template: %w", err)
		}

		if len(renderedMessages) != 1 || renderedMessages[0] == nil {
			return "", fmt.Errorf("rendered react classification prompt messages = %d, want 1", len(renderedMessages))
		}

		return strings.TrimSpace(renderedMessages[0].Content), nil
	}
}

func reactPromptExclusionaryList(entries []*entities.ExclusionListEntry) []map[string]string {
	items := make([]map[string]string, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}

		items = append(items, map[string]string{
			"activity_type": strings.TrimSpace(entry.ActivityType()),
		})
	}

	return items
}

func reactPromptU1List(entries []*entities.U1ListEntry) []map[string]string {
	items := make([]map[string]string, 0, len(entries))
	for _, entry := range entries {
		if entry == nil {
			continue
		}

		items = append(items, map[string]string{
			"sector":                  strings.TrimSpace(entry.Sector()),
			"eligible_operation_type": strings.TrimSpace(entry.EligibleOperationType()),
			"condition_guidance":      strings.TrimSpace(entry.ConditionGuidance()),
		})
	}

	return items
}
