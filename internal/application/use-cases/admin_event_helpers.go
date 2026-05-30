package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/gyud-adb/paris-api/internal/domain"
	"github.com/gyud-adb/paris-api/internal/domain/events"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

const (
	createUserAdminEventType                           = events.CreateUserEventType
	getUserAdminEventType                              = events.GetUserEventType
	listUsersAdminEventType                            = events.ListUsersEventType
	updateUserAdminEventType                           = events.UpdateUserEventType
	deleteUserAdminEventType                           = events.DeleteUserEventType
	createGroupAdminEventType                          = events.CreateGroupEventType
	getGroupAdminEventType                             = events.GetGroupEventType
	listGroupsAdminEventType                           = events.ListGroupsEventType
	updateGroupAdminEventType                          = events.UpdateGroupEventType
	deleteGroupAdminEventType                          = events.DeleteGroupEventType
	createU1ListAdminEventType                         = events.CreateU1ListEventType
	getU1ListAdminEventType                            = events.GetU1ListEventType
	listU1ListAdminEventType                           = events.ListU1ListEventType
	updateU1ListAdminEventType                         = events.UpdateU1ListEventType
	deleteU1ListAdminEventType                         = events.DeleteU1ListEventType
	createExclusionListAdminEventType                  = events.CreateExclusionListEventType
	getExclusionListAdminEventType                     = events.GetExclusionListEventType
	listExclusionListAdminEventType                    = events.ListExclusionListEventType
	updateExclusionListAdminEventType                  = events.UpdateExclusionListEventType
	deleteExclusionListAdminEventType                  = events.DeleteExclusionListEventType
	createSectorAdminEventType                         = events.CreateSectorEventType
	getSectorAdminEventType                            = events.GetSectorEventType
	listSectorsAdminEventType                          = events.ListSectorsEventType
	updateSectorAdminEventType                         = events.UpdateSectorEventType
	deleteSectorAdminEventType                         = events.DeleteSectorEventType
	createTransactionAdminEventType                    = events.CreateTransactionEventType
	getTransactionAdminEventType                       = events.GetTransactionEventType
	listTransactionsAdminEventType                     = events.ListTransactionsEventType
	deleteTransactionAdminEventType                    = events.DeleteTransactionEventType
	getTransactionUploadAdminEventType                 = events.GetTransactionUploadEventType
	createTransactionUploadAdminEventType              = events.CreateTransactionUploadEventType
	deleteTransactionUploadAdminEventType              = events.DeleteTransactionUploadEventType
	createBugReportAdminEventType                      = events.CreateBugReportEventType
	getBugReportAdminEventType                         = events.GetBugReportEventType
	listBugReportsAdminEventType                       = events.ListBugReportsEventType
	updateBugReportAdminEventType                      = events.UpdateBugReportEventType
	deleteBugReportAdminEventType                      = events.DeleteBugReportEventType
	createTransactionStep4AdminEventType               = events.CreateTransactionStep4EventType
	createTransactionStep5AdminEventType               = events.CreateTransactionStep5EventType
	retryTransactionUploadClassificationAdminEventType = events.RetryTransactionUploadClassificationEventType
	upsertTransactionFeedbackAdminEventType            = events.UpsertTransactionFeedbackEventType
	getTransactionFeedbackAdminEventType               = events.GetTransactionFeedbackEventType
	deleteTransactionFeedbackAdminEventType            = events.DeleteTransactionFeedbackEventType
)

type adminEventRecorder = ports.EventPublisher

func recordAdminEvent(ctx context.Context, publisher ports.EventPublisher, occurredAt time.Time, actorUserID, actorGroupID, eventType string, payload map[string]any) error {
	event, err := events.NewAdminActionOccurred(occurredAt, actorUserID, actorGroupID, eventType, payload)
	if err != nil {
		return fmt.Errorf("creating admin event: %w", err)
	}

	if err := publisher.Publish(ctx, []domain.DomainEvent{event}); err != nil {
		return fmt.Errorf("publishing admin event: %w", err)
	}

	return nil
}
