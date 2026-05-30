package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	"go.uber.org/zap"
)

const step3InvalidTenorReason = "tenor description missing or unrecognized; defaulting step 3 result to not aligned"

// ReActClassificationJobHandler dequeues and persists ReAct transaction classification work.
type ReActClassificationJobHandler struct {
	gateway outboundports.TransactionClassificationGateway
	txRepo  outboundports.TransactionRepository
	logger  *zap.Logger
	now     func() time.Time
}

type reactClassificationHandledError struct {
	cause            error
	requiresRollback bool
}

type transactionNotFoundError struct {
	transactionID string
}

// Error reports the missing transaction in a worker-friendly format.
func (e *transactionNotFoundError) Error() string {
	return fmt.Sprintf("transaction %s was not found", e.transactionID)
}

// Error returns the underlying classification failure message when available.
func (e *reactClassificationHandledError) Error() string {
	if e == nil || e.cause == nil {
		return "react classification failed"
	}

	return e.cause.Error()
}

// Unwrap exposes the underlying failure so callers can inspect or match it.
func (e *reactClassificationHandledError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.cause
}

// NewReActClassificationJobHandler builds a ReActClassificationJobHandler.
func NewReActClassificationJobHandler(
	gateway outboundports.TransactionClassificationGateway,
	txRepo outboundports.TransactionRepository,
	logger *zap.Logger,
) *ReActClassificationJobHandler {
	return &ReActClassificationJobHandler{
		gateway: gateway,
		txRepo:  txRepo,
		logger:  logger,
		now:     time.Now,
	}
}

// Handle executes one queued ReAct transaction classification job.
func (h *ReActClassificationJobHandler) Handle(ctx context.Context, job outboundports.ClassificationJob) error {
	return h.HandleBatch(ctx, []outboundports.ClassificationJob{job})
}

// HandleBatch executes a batch of queued ReAct transaction classification jobs.
func (h *ReActClassificationJobHandler) HandleBatch(ctx context.Context, jobs []outboundports.ClassificationJob) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(jobs) == 0 {
		return nil
	}

	transactions := make([]*entities.Transaction, 0, len(jobs))
	candidates := make([]outboundports.TransactionClassificationCandidate, 0, len(jobs))
	transactionByID := make(map[string]*entities.Transaction, len(jobs))
	for _, job := range jobs {
		transactionID, err := valueobjects.TransactionIDFromString(job.Payload.TransactionID)
		if err != nil {
			return fmt.Errorf("parsing transaction id from react classification job: %w", err)
		}

		transaction, err := h.txRepo.FindByID(ctx, transactionID)
		if err != nil {
			return fmt.Errorf("loading transaction %s for react classification: %w", transactionID.String(), err)
		}

		if transaction == nil {
			return &transactionNotFoundError{transactionID: transactionID.String()}
		}

		if h.shouldSkipCompleted(transaction) {
			continue
		}

		if transaction.Status() != valueobjects.ProcessingTransactionStatus().String() {
			h.warn("skipping react classification job for transaction with unexpected status",
				zap.String("transaction_id", transaction.ID().String()),
				zap.String("status", transaction.Status()),
			)
			continue
		}

		transactions = append(transactions, transaction)
		candidates = append(candidates, newTransactionClassificationCandidate(transaction))
		transactionByID[transaction.ID().String()] = transaction
	}

	if len(candidates) == 0 {
		return nil
	}

	decisions, err := h.gateway.Classify(ctx, candidates)
	if err != nil {
		if failErr := h.failTransactions(ctx, transactions, err); failErr != nil {
			return failErr
		}

		return &reactClassificationHandledError{cause: err}
	}

	if len(decisions) != len(candidates) {
		classificationErr := fmt.Errorf("react classification gateway returned %d decisions, want %d", len(decisions), len(candidates))
		if failErr := h.failTransactions(ctx, transactions, classificationErr); failErr != nil {
			return failErr
		}

		return &reactClassificationHandledError{cause: classificationErr}
	}

	orderedIDs := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		orderedIDs = append(orderedIDs, candidate.TransactionID.String())
	}

	decisionByID := make(map[string]outboundports.TransactionClassificationDecision, len(decisions))
	for _, decision := range decisions {
		decisionByID[decision.TransactionID.String()] = decision
	}

	for _, transactionID := range orderedIDs {
		transaction := transactionByID[transactionID]
		decision, ok := decisionByID[transactionID]
		if !ok {
			classificationErr := fmt.Errorf("react classification decision missing for transaction %s", transactionID)
			if failErr := h.failTransactions(ctx, transactions, classificationErr); failErr != nil {
				return failErr
			}

			return &reactClassificationHandledError{cause: classificationErr}
		}

		resolvedDecision, err := h.applyDecision(ctx, transaction, decision)
		if err != nil {
			classificationErr := fmt.Errorf("applying react classification decision for transaction %s: %w", transaction.ID().String(), err)
			if failErr := h.failTransactions(ctx, transactions, classificationErr); failErr != nil {
				return failErr
			}

			return &reactClassificationHandledError{cause: classificationErr}
		}

		if err := h.txRepo.Update(ctx, transaction); err != nil {
			return fmt.Errorf("persisting react-classified transaction %s: %w", transaction.ID().String(), err)
		}

		h.info("react classification job completed successfully",
			zap.String("transaction_id", transaction.ID().String()),
			zap.String("final_classification", resolvedDecision.Classification.String()),
			zap.String("status", resolvedDecision.Status.String()),
			zap.String("source", resolvedDecision.Source),
		)
	}

	return nil
}

func (h *ReActClassificationJobHandler) shouldSkipCompleted(transaction *entities.Transaction) bool {
	status := transaction.Status()
	if status != valueobjects.AIReviewedTransactionStatus().String() &&
		status != valueobjects.FromPreviousTransactionsTransactionStatus().String() &&
		status != valueobjects.ProfessionallyReviewedTransactionStatus().String() &&
		status != valueobjects.FailedTransactionStatus().String() {
		return false
	}

	h.warn("skipping react classification job for terminal transaction",
		zap.String("transaction_id", transaction.ID().String()),
		zap.String("status", status),
	)

	return true
}

func (h *ReActClassificationJobHandler) markFailed(ctx context.Context, transaction *entities.Transaction, classificationErr error) error {
	failureReason := strings.TrimSpace(classificationErr.Error())
	if err := transaction.MarkFailed(failureReason, h.now()); err != nil {
		return fmt.Errorf("marking transaction failed: %w", err)
	}

	if err := h.txRepo.Update(ctx, transaction); err != nil {
		return fmt.Errorf("persisting failed transaction: %w", err)
	}

	return nil
}

func (h *ReActClassificationJobHandler) failTransactions(ctx context.Context, transactions []*entities.Transaction, classificationErr error) error {
	for _, transaction := range transactions {
		if err := h.markFailed(ctx, transaction, classificationErr); err != nil {
			if requiresRollbackAfterFailurePersistence(err) {
				return &reactClassificationHandledError{cause: err, requiresRollback: true}
			}

			return fmt.Errorf("marking transaction %s failed after react classification error: %w", transaction.ID().String(), err)
		}
	}

	return nil
}

func requiresRollbackAfterFailurePersistence(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "SQLSTATE 25P02")
}

func (h *ReActClassificationJobHandler) applyDecision(ctx context.Context, transaction *entities.Transaction, decision outboundports.TransactionClassificationDecision) (outboundports.TransactionClassificationDecision, error) {
	fallbackDecision, err := h.resolveStep3Fallback(ctx, transaction, decision)
	if err != nil {
		return outboundports.TransactionClassificationDecision{}, err
	}

	decision = fallbackDecision

	switch {
	case decision.Status.Equal(valueobjects.FromPreviousTransactionsTransactionStatus()):
		if err := transaction.MarkClassifiedFromPreviousTransaction(decision.Classification, decision.ReviewResult, h.now()); err != nil {
			return outboundports.TransactionClassificationDecision{}, err
		}
		return decision, nil
	case decision.Status.Equal(valueobjects.AIReviewedTransactionStatus()):
		if err := transaction.MarkClassified(decision.Classification, decision.ReviewResult, h.now()); err != nil {
			return outboundports.TransactionClassificationDecision{}, err
		}
		return decision, nil
	default:
		return outboundports.TransactionClassificationDecision{}, fmt.Errorf("unsupported react classification status %q", decision.Status.String())
	}
}

func (h *ReActClassificationJobHandler) resolveStep3Fallback(ctx context.Context, transaction *entities.Transaction, decision outboundports.TransactionClassificationDecision) (outboundports.TransactionClassificationDecision, error) {
	if !shouldFallbackToStep3(decision) {
		return decision, nil
	}

	step3Result, err := executeStep3Rule(ctx, transaction)
	if err != nil {
		return outboundports.TransactionClassificationDecision{}, err
	}

	react := decision.ReviewResult.React()
	if react == nil {
		return outboundports.TransactionClassificationDecision{}, fmt.Errorf("react review result is required for step 3 fallback")
	}

	finalClassification := valueobjects.NextStepTransactionClassification()
	if booleanResult := step3Result.BooleanResult(); booleanResult != nil && *booleanResult {
		finalClassification = valueobjects.AlignedTransactionClassification()
	}

	combinedReason := strings.TrimSpace(react.Reason())
	if step3Reason := step3FallbackReason(step3Result); step3Reason != "" {
		if combinedReason == "" {
			combinedReason = step3Reason
		} else {
			combinedReason = combinedReason + "; " + step3Reason
		}
	}

	updatedReact := valueobjects.NewReactReviewResult(
		valueobjects.PipelineResultSourceLegacyStep3Fallback,
		react.ClassifierFamily(),
		react.ClassifierVersion(),
		react.PromptVersion(),
		react.Model(),
		react.TransactionID(),
		react.MatchedTransactionID(),
		react.MatchedGoodsDescription(),
		react.BatchID(),
		react.BatchSize(),
		react.NotAlignedListMatch(),
		react.NotAlignedListMatchConfidence(),
		react.AlignedListMatch(),
		react.AlignedListMatchConfidence(),
		3,
		finalClassification,
		combinedReason,
	)

	updatedDecision := decision
	updatedDecision.Classification = finalClassification
	updatedDecision.Status = valueobjects.AIReviewedTransactionStatus()
	updatedDecision.ReviewResult = valueobjects.NewReactPipelineResult(updatedReact)
	updatedDecision.Source = valueobjects.PipelineResultSourceLegacyStep3Fallback
	return updatedDecision, nil
}
func executeStep3Rule(ctx context.Context, transaction *entities.Transaction) (valueobjects.StepResult, error) {
	if transaction == nil {
		return valueobjects.StepResult{}, fmt.Errorf("transaction is required")
	}

	_ = ctx

	normalizedTenor := strings.ToUpper(strings.TrimSpace(transaction.TenorDescription()))
	switch {
	case normalizedTenor == "N":
		return valueobjects.NewBooleanStepResult(3, true), nil
	case strings.HasPrefix(normalizedTenor, "Y"):
		return valueobjects.NewBooleanStepResult(3, false), nil
	default:
		return valueobjects.NewBooleanStepResultWithReason(3, false, errors.New(step3InvalidTenorReason)), nil
	}
}

func shouldFallbackToStep3(decision outboundports.TransactionClassificationDecision) bool {
	if !decision.Status.Equal(valueobjects.AIReviewedTransactionStatus()) {
		return false
	}

	react := decision.ReviewResult.React()
	if react == nil {
		return false
	}

	if decision.Classification.Equal(valueobjects.AlignedTransactionClassification()) || decision.Classification.Equal(valueobjects.NotAlignedTransactionClassification()) {
		return false
	}

	if react.ExitStep() == 0 {
		return true
	}

	return react.ExitStep() >= 2 && decision.Classification.Equal(valueobjects.NextStepTransactionClassification())
}

func step3FallbackReason(result valueobjects.StepResult) string {
	booleanResult := result.BooleanResult()
	if booleanResult == nil {
		return "legacy step 3 fallback produced no boolean result"
	}

	outcome := "false"
	if *booleanResult {
		outcome = "true"
	}

	reason := fmt.Sprintf("legacy step 3 fallback boolean result: %s", outcome)
	if resultReason := result.Reason(); resultReason != nil {
		reason = reason + " (" + strings.TrimSpace(resultReason.Error()) + ")"
	}

	return reason
}

func newTransactionClassificationCandidate(transaction *entities.Transaction) outboundports.TransactionClassificationCandidate {
	return outboundports.TransactionClassificationCandidate{
		TransactionID:    transaction.ID(),
		GoodsDescription: transaction.GoodsDescription(),
	}
}

func (h *ReActClassificationJobHandler) warn(message string, fields ...zap.Field) {
	if h == nil || h.logger == nil {
		return
	}

	h.logger.Warn(message, fields...)
}

func (h *ReActClassificationJobHandler) info(message string, fields ...zap.Field) {
	if h == nil || h.logger == nil {
		return
	}

	h.logger.Info(message, fields...)
}
