package usecases

import (
	"context"
	"fmt"
	"time"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"go.uber.org/zap"
)

const retryTransactionUploadClassificationTask = outboundports.TransactionClassifyReactTaskName

// RetryTransactionUploadClassificationUseCase manually retries failed upload classifications.
type RetryTransactionUploadClassificationUseCase struct {
	uploadRepository   outboundports.TransactionUploadRepository
	retryRepository    outboundports.TransactionClassificationRetryRepository
	transactionManager outboundports.TransactionManager
	eventRecorder      adminEventRecorder
	actorDirectory     outboundports.ActorDirectory
	logger             *zap.Logger
	now                func() time.Time
}

// NewRetryTransactionUploadClassificationUseCase builds a RetryTransactionUploadClassificationUseCase.
func NewRetryTransactionUploadClassificationUseCase(
	uploadRepository outboundports.TransactionUploadRepository,
	retryRepository outboundports.TransactionClassificationRetryRepository,
	transactionManager outboundports.TransactionManager,
	eventRecorder adminEventRecorder,
	actorDirectory outboundports.ActorDirectory,
	logger *zap.Logger,
) *RetryTransactionUploadClassificationUseCase {
	return &RetryTransactionUploadClassificationUseCase{
		uploadRepository:   uploadRepository,
		retryRepository:    retryRepository,
		transactionManager: transactionManager,
		eventRecorder:      eventRecorder,
		actorDirectory:     actorDirectory,
		logger:             logger,
		now:                time.Now,
	}
}

// Execute retries failed classifications for one upload.
func (uc *RetryTransactionUploadClassificationUseCase) Execute(ctx context.Context, command inboundports.RetryTransactionUploadClassificationCommand) (inboundports.RetryTransactionUploadClassificationResult, error) {
	if err := validateActor(ctx, uc.actorDirectory, command.ActorUserID, command.ActorGroupID); err != nil {
		return inboundports.RetryTransactionUploadClassificationResult{}, err
	}

	uploadID, err := valueobjects.UploadIDFromString(command.UploadID)
	if err != nil {
		return inboundports.RetryTransactionUploadClassificationResult{}, fmt.Errorf("parsing upload id: %w", err)
	}

	upload, err := uc.uploadRepository.FindByID(ctx, uploadID)
	if err != nil {
		return inboundports.RetryTransactionUploadClassificationResult{}, fmt.Errorf("finding upload by id: %w", err)
	}

	if upload == nil {
		uc.logRetryNotFound(command)
		if err := uc.recordRetryNotFoundAuditEvent(ctx, command); err != nil {
			return inboundports.RetryTransactionUploadClassificationResult{}, err
		}

		return inboundports.RetryTransactionUploadClassificationResult{}, &NotFoundError{Resource: "transaction upload", ID: command.UploadID}
	}

	failedTransactionIDs, err := uc.retryRepository.ListFailedByUploadID(ctx, uploadID)
	if err != nil {
		return inboundports.RetryTransactionUploadClassificationResult{}, fmt.Errorf("listing failed transactions by upload: %w", err)
	}

	result := inboundports.RetryTransactionUploadClassificationResult{
		UploadID:                   command.UploadID,
		EligibleFailedTransactions: len(failedTransactionIDs),
	}

	retriedAt := uc.now()
	for _, transactionID := range failedTransactionIDs {
		var retryResult outboundports.RetryFailedTransactionResult
		err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
			createdResult, retryErr := uc.retryRepository.RetryFailedTransaction(txCtx, outboundports.RetryFailedTransactionCommand{
				UploadID:      uploadID,
				TransactionID: transactionID,
				TaskName:      retryTransactionUploadClassificationTask,
				LastRetriedAt: retriedAt,
			})
			if retryErr != nil {
				return retryErr
			}

			retryResult = createdResult
			return nil
		})
		if err != nil {
			result.FailedRetryCreations++
			result.Failures = append(result.Failures, inboundports.RetryTransactionUploadClassificationFailure{
				TransactionID: transactionID.String(),
				Error:         retryFailureMessage(err),
			})
			continue
		}

		if retryResult.Skipped {
			result.SkippedTransactions++
			result.Skipped = append(result.Skipped, inboundports.RetryTransactionUploadClassificationSkippedTransaction{
				TransactionID: transactionID.String(),
				Reason:        retrySkipReason(retryResult.SkipReason),
			})
			continue
		}

		if retryResult.Attempt == nil {
			result.FailedRetryCreations++
			result.Failures = append(result.Failures, inboundports.RetryTransactionUploadClassificationFailure{
				TransactionID: transactionID.String(),
				Error:         "retry attempt was not created",
			})
			continue
		}

		result.RetriedTransactions++
	}

	if err := uc.recordRetryAuditEvent(ctx, command, result); err != nil {
		return inboundports.RetryTransactionUploadClassificationResult{}, err
	}

	uc.logRetryResult(command, result)
	return result, nil
}

func (uc *RetryTransactionUploadClassificationUseCase) recordRetryAuditEvent(ctx context.Context, command inboundports.RetryTransactionUploadClassificationCommand, result inboundports.RetryTransactionUploadClassificationResult) error {
	failures := make([]map[string]any, 0, len(result.Failures))
	for _, failure := range result.Failures {
		failures = append(failures, map[string]any{
			"transaction_id": failure.TransactionID,
			"error":          failure.Error,
		})
	}

	payload := map[string]any{
		"action":                       "retry",
		"resource":                     "transaction_upload_classification",
		"upload_id":                    result.UploadID,
		"eligible_failed_transactions": result.EligibleFailedTransactions,
		"retried_transactions":         result.RetriedTransactions,
		"skipped_transactions":         result.SkippedTransactions,
		"skipped":                      skippedTransactionsPayload(result.Skipped),
		"failed_retry_creations":       result.FailedRetryCreations,
		"outcome":                      retryTransactionUploadClassificationOutcome(result),
		"failures":                     failures,
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), command.ActorUserID, command.ActorGroupID, retryTransactionUploadClassificationAdminEventType, payload); err != nil {
		return fmt.Errorf("recording transaction upload classification retry event: %w", err)
	}

	return nil
}

func (uc *RetryTransactionUploadClassificationUseCase) recordRetryNotFoundAuditEvent(ctx context.Context, command inboundports.RetryTransactionUploadClassificationCommand) error {
	payload := map[string]any{
		"action":                       "retry",
		"resource":                     "transaction_upload_classification",
		"upload_id":                    command.UploadID,
		"eligible_failed_transactions": 0,
		"retried_transactions":         0,
		"skipped_transactions":         0,
		"failed_retry_creations":       0,
		"outcome":                      "not_found",
		"failures":                     []map[string]any{},
	}

	if err := recordAdminEvent(ctx, uc.eventRecorder, uc.now(), command.ActorUserID, command.ActorGroupID, retryTransactionUploadClassificationAdminEventType, payload); err != nil {
		return fmt.Errorf("recording transaction upload classification retry event: %w", err)
	}

	return nil
}

func (uc *RetryTransactionUploadClassificationUseCase) logRetryResult(command inboundports.RetryTransactionUploadClassificationCommand, result inboundports.RetryTransactionUploadClassificationResult) {
	if uc == nil || uc.logger == nil {
		return
	}

	fields := []zap.Field{
		zap.String("upload_id", result.UploadID),
		zap.String("actor_user_id", command.ActorUserID),
		zap.String("actor_group_id", command.ActorGroupID),
		zap.Int("eligible_count", result.EligibleFailedTransactions),
		zap.Int("retried_count", result.RetriedTransactions),
		zap.Int("skipped_count", result.SkippedTransactions),
		zap.Any("skipped", result.Skipped),
		zap.Int("failed_retry_creations", result.FailedRetryCreations),
		zap.String("outcome", retryTransactionUploadClassificationOutcome(result)),
	}

	if len(result.Failures) > 0 {
		fields = append(fields, zap.Any("failures", result.Failures))
		uc.logger.Warn("transaction upload classification retry completed with failures", fields...)
		return
	}

	uc.logger.Info("transaction upload classification retry completed", fields...)
}

func (uc *RetryTransactionUploadClassificationUseCase) logRetryNotFound(command inboundports.RetryTransactionUploadClassificationCommand) {
	if uc == nil || uc.logger == nil {
		return
	}

	uc.logger.Warn(
		"transaction upload classification retry upload not found",
		zap.String("upload_id", command.UploadID),
		zap.String("actor_user_id", command.ActorUserID),
		zap.String("actor_group_id", command.ActorGroupID),
		zap.String("outcome", "not_found"),
	)
}

func retryTransactionUploadClassificationOutcome(result inboundports.RetryTransactionUploadClassificationResult) string {
	if result.EligibleFailedTransactions == 0 {
		return "no_op"
	}

	if result.FailedRetryCreations > 0 {
		return "partial_failure"
	}

	if result.RetriedTransactions == 0 && result.SkippedTransactions > 0 {
		return "skipped"
	}

	return "completed"
}

func retryFailureMessage(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func retrySkipReason(reason string) string {
	if reason == "" {
		return "already_queued_or_not_failed"
	}

	return reason
}

func skippedTransactionsPayload(skipped []inboundports.RetryTransactionUploadClassificationSkippedTransaction) []map[string]any {
	payload := make([]map[string]any, 0, len(skipped))
	for _, item := range skipped {
		payload = append(payload, map[string]any{
			"transaction_id": item.TransactionID,
			"reason":         item.Reason,
		})
	}

	return payload
}
