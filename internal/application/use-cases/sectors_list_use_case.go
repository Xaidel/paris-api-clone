package usecases

import (
	"context"
	"fmt"
	"time"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// ListSectorsUseCase lists sector entries.
type ListSectorsUseCase struct {
	repository    outboundports.SectorRepository
	eventRecorder adminEventRecorder
	now           func() time.Time
}

// NewListSectorsUseCase builds a ListSectorsUseCase.
func NewListSectorsUseCase(repository outboundports.SectorRepository, eventRecorder adminEventRecorder) *ListSectorsUseCase {
	return &ListSectorsUseCase{repository: repository, eventRecorder: eventRecorder, now: time.Now}
}

// Execute lists all sector entries.
func (uc *ListSectorsUseCase) Execute(ctx context.Context, query inboundports.ListSectorsQuery) (outboundports.ListSectorsResult, error) {
	sectors, err := uc.repository.List(ctx)
	if err != nil {
		return outboundports.ListSectorsResult{}, fmt.Errorf("listing sector entries: %w", err)
	}

	results := make([]outboundports.SectorResult, 0, len(sectors))
	for _, sector := range sectors {
		results = append(results, newSectorResult(sector))
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), query.ActorUserID, query.ActorGroupID, listSectorsAdminEventType, map[string]any{
		"action":       "list",
		"resource":     "sector",
		"result_count": len(results),
	}); err != nil {
		return outboundports.ListSectorsResult{}, err
	}

	return outboundports.ListSectorsResult{Sectors: results}, nil
}
