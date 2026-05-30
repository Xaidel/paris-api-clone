package usecases

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domain "github.com/gyud-adb/paris-api/internal/domain"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type transactionClassificationRetryRepositoryMock struct {
	listResult []valueobjects.TransactionID
	listErr    error

	retryAttempts []*outboundports.RetryFailedTransactionCommand
	retryResults  []outboundports.RetryFailedTransactionResult
	retryErrs     []error
	callCount     int
}

func (m *transactionClassificationRetryRepositoryMock) ListFailedByUploadID(_ context.Context, _ valueobjects.UploadID) ([]valueobjects.TransactionID, error) {
	return m.listResult, m.listErr
}

func (m *transactionClassificationRetryRepositoryMock) RetryFailedTransaction(_ context.Context, command outboundports.RetryFailedTransactionCommand) (outboundports.RetryFailedTransactionResult, error) {
	m.retryAttempts = append(m.retryAttempts, &command)
	index := m.callCount
	m.callCount++

	if index < len(m.retryErrs) && m.retryErrs[index] != nil {
		return outboundports.RetryFailedTransactionResult{}, m.retryErrs[index]
	}

	if index < len(m.retryResults) {
		return m.retryResults[index], nil
	}

	return outboundports.RetryFailedTransactionResult{}, nil
}

func TestRetryTransactionUploadClassificationUseCaseExecute(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transactionID1, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transactionID2, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	tests := []struct {
		name        string
		uploadRepo  *transactionUploadRepositoryMock
		retryRepo   *transactionClassificationRetryRepositoryMock
		manager     *transactionManagerUseCaseMock
		recorder    *adminEventRecorderUploadMock
		actors      *actorDirectoryMock
		command     inboundports.RetryTransactionUploadClassificationCommand
		assertError func(t *testing.T, err error)
		assert      func(t *testing.T, result inboundports.RetryTransactionUploadClassificationResult, retryRepo *transactionClassificationRetryRepositoryMock, manager *transactionManagerUseCaseMock, recorder *adminEventRecorderUploadMock)
	}{
		{
			name:       "retries failed upload transactions",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, valueobjects.UploadedTransactionUploadStatus(), 2)},
			retryRepo: &transactionClassificationRetryRepositoryMock{
				listResult: []valueobjects.TransactionID{transactionID1, transactionID2},
				retryResults: []outboundports.RetryFailedTransactionResult{
					{Attempt: &outboundports.TransactionClassificationRetryAttempt{JobID: "job-1", RetryCount: 1, LastRetriedAt: testTime()}},
					{Skipped: true, SkipReason: "already_queued_or_not_failed"},
				},
			},
			manager:  &transactionManagerUseCaseMock{},
			recorder: &adminEventRecorderUploadMock{},
			actors:   &actorDirectoryMock{},
			command:  inboundports.RetryTransactionUploadClassificationCommand{UploadID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result inboundports.RetryTransactionUploadClassificationResult, retryRepo *transactionClassificationRetryRepositoryMock, manager *transactionManagerUseCaseMock, recorder *adminEventRecorderUploadMock) {
				t.Helper()

				if result.EligibleFailedTransactions != 2 {
					t.Fatalf("result.EligibleFailedTransactions = %d, want %d", result.EligibleFailedTransactions, 2)
				}
				if result.RetriedTransactions != 1 {
					t.Fatalf("result.RetriedTransactions = %d, want %d", result.RetriedTransactions, 1)
				}
				if result.SkippedTransactions != 1 {
					t.Fatalf("result.SkippedTransactions = %d, want %d", result.SkippedTransactions, 1)
				}
				if len(result.Skipped) != 1 {
					t.Fatalf("len(result.Skipped) = %d, want %d", len(result.Skipped), 1)
				}
				if result.Skipped[0].Reason != "already_queued_or_not_failed" {
					t.Fatalf("result.Skipped[0].Reason = %q, want %q", result.Skipped[0].Reason, "already_queued_or_not_failed")
				}
				if result.FailedRetryCreations != 0 {
					t.Fatalf("result.FailedRetryCreations = %d, want %d", result.FailedRetryCreations, 0)
				}
				if !manager.invoked {
					t.Fatal("expected transaction manager to be invoked")
				}
				if len(retryRepo.retryAttempts) != 2 {
					t.Fatalf("len(retryRepo.retryAttempts) = %d, want %d", len(retryRepo.retryAttempts), 2)
				}
				if retryRepo.retryAttempts[0].TaskName != outboundports.TransactionClassifyReactTaskName {
					t.Fatalf("retryRepo.retryAttempts[0].TaskName = %q, want %q", retryRepo.retryAttempts[0].TaskName, outboundports.TransactionClassifyReactTaskName)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				if payload["outcome"] != "completed" {
					t.Fatalf("payload[outcome] = %v, want %q", payload["outcome"], "completed")
				}
			},
		},
		{
			name:       "returns no-op when upload has no failed transactions",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, valueobjects.FailedTransactionUploadStatus(), 0)},
			retryRepo:  &transactionClassificationRetryRepositoryMock{},
			manager:    &transactionManagerUseCaseMock{},
			recorder:   &adminEventRecorderUploadMock{},
			actors:     &actorDirectoryMock{},
			command:    inboundports.RetryTransactionUploadClassificationCommand{UploadID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result inboundports.RetryTransactionUploadClassificationResult, retryRepo *transactionClassificationRetryRepositoryMock, manager *transactionManagerUseCaseMock, recorder *adminEventRecorderUploadMock) {
				t.Helper()

				if result.EligibleFailedTransactions != 0 {
					t.Fatalf("result.EligibleFailedTransactions = %d, want %d", result.EligibleFailedTransactions, 0)
				}
				if len(retryRepo.retryAttempts) != 0 {
					t.Fatalf("len(retryRepo.retryAttempts) = %d, want %d", len(retryRepo.retryAttempts), 0)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				if payload["outcome"] != "no_op" {
					t.Fatalf("payload[outcome] = %v, want %q", payload["outcome"], "no_op")
				}
			},
		},
		{
			name:       "returns not found for missing upload",
			uploadRepo: &transactionUploadRepositoryMock{},
			retryRepo:  &transactionClassificationRetryRepositoryMock{},
			manager:    &transactionManagerUseCaseMock{},
			recorder:   &adminEventRecorderUploadMock{},
			actors:     &actorDirectoryMock{},
			command:    inboundports.RetryTransactionUploadClassificationCommand{UploadID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assertError: func(t *testing.T, err error) {
				t.Helper()

				var notFoundErr *NotFoundError
				if !errors.As(err, &notFoundErr) {
					t.Fatalf("expected NotFoundError, got %v", err)
				}
			},
			assert: func(t *testing.T, _ inboundports.RetryTransactionUploadClassificationResult, _ *transactionClassificationRetryRepositoryMock, _ *transactionManagerUseCaseMock, recorder *adminEventRecorderUploadMock) {
				t.Helper()

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				if payload["outcome"] != "not_found" {
					t.Fatalf("payload[outcome] = %v, want %q", payload["outcome"], "not_found")
				}
			},
		},
		{
			name:       "records partial failures",
			uploadRepo: &transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, valueobjects.UploadedTransactionUploadStatus(), 2)},
			retryRepo: &transactionClassificationRetryRepositoryMock{
				listResult: []valueobjects.TransactionID{transactionID1, transactionID2},
				retryResults: []outboundports.RetryFailedTransactionResult{
					{Attempt: &outboundports.TransactionClassificationRetryAttempt{JobID: "job-1", RetryCount: 1, LastRetriedAt: testTime()}},
				},
				retryErrs: []error{nil, errors.New("boom")},
			},
			manager:  &transactionManagerUseCaseMock{},
			recorder: &adminEventRecorderUploadMock{},
			actors:   &actorDirectoryMock{},
			command:  inboundports.RetryTransactionUploadClassificationCommand{UploadID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"},
			assert: func(t *testing.T, result inboundports.RetryTransactionUploadClassificationResult, _ *transactionClassificationRetryRepositoryMock, _ *transactionManagerUseCaseMock, recorder *adminEventRecorderUploadMock) {
				t.Helper()

				if result.FailedRetryCreations != 1 {
					t.Fatalf("result.FailedRetryCreations = %d, want %d", result.FailedRetryCreations, 1)
				}
				if len(result.Failures) != 1 {
					t.Fatalf("len(result.Failures) = %d, want %d", len(result.Failures), 1)
				}

				var payload map[string]any
				if err := json.Unmarshal(recorder.command.EventData, &payload); err != nil {
					t.Fatalf("json.Unmarshal() error = %v", err)
				}
				if payload["outcome"] != "partial_failure" {
					t.Fatalf("payload[outcome] = %v, want %q", payload["outcome"], "partial_failure")
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			useCase := NewRetryTransactionUploadClassificationUseCase(tc.uploadRepo, tc.retryRepo, tc.manager, tc.recorder, tc.actors, zap.NewNop())
			useCase.now = testTime

			result, err := useCase.Execute(context.Background(), tc.command)
			if tc.assertError != nil {
				tc.assertError(t, err)
				if tc.assert != nil {
					tc.assert(t, result, tc.retryRepo, tc.manager, tc.recorder)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			tc.assert(t, result, tc.retryRepo, tc.manager, tc.recorder)
		})
	}
}

func TestRetryTransactionUploadClassificationUseCaseExecuteAuditFailure(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	useCase := NewRetryTransactionUploadClassificationUseCase(
		&transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, valueobjects.FailedTransactionUploadStatus(), 0)},
		&transactionClassificationRetryRepositoryMock{},
		&transactionManagerUseCaseMock{},
		&adminEventRecorderUploadMock{err: errors.New("audit failed")},
		&actorDirectoryMock{},
		zap.NewNop(),
	)
	useCase.now = testTime

	_, err = useCase.Execute(context.Background(), inboundports.RetryTransactionUploadClassificationCommand{UploadID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"})
	if err == nil || err.Error() != "recording transaction upload classification retry event: publishing admin event: audit failed" {
		t.Fatalf("Execute() error = %v, want %q", err, "recording transaction upload classification retry event: publishing admin event: audit failed")
	}
}

func TestRetryTransactionUploadClassificationOutcome(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result inboundports.RetryTransactionUploadClassificationResult
		want   string
	}{
		{name: "no op", result: inboundports.RetryTransactionUploadClassificationResult{}, want: "no_op"},
		{name: "partial failure", result: inboundports.RetryTransactionUploadClassificationResult{EligibleFailedTransactions: 2, FailedRetryCreations: 1}, want: "partial_failure"},
		{name: "skipped", result: inboundports.RetryTransactionUploadClassificationResult{EligibleFailedTransactions: 1, SkippedTransactions: 1}, want: "skipped"},
		{name: "completed", result: inboundports.RetryTransactionUploadClassificationResult{EligibleFailedTransactions: 2, RetriedTransactions: 2}, want: "completed"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := retryTransactionUploadClassificationOutcome(tc.result); got != tc.want {
				t.Fatalf("retryTransactionUploadClassificationOutcome() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRetryTransactionUploadClassificationUseCaseExecutePropagatesActorValidation(t *testing.T) {
	t.Parallel()

	useCase := NewRetryTransactionUploadClassificationUseCase(
		&transactionUploadRepositoryMock{},
		&transactionClassificationRetryRepositoryMock{},
		&transactionManagerUseCaseMock{},
		&adminEventRecorderUploadMock{},
		&actorDirectoryMock{err: domain.ErrUnknownActorUserID},
		zap.NewNop(),
	)

	_, err := useCase.Execute(context.Background(), inboundports.RetryTransactionUploadClassificationCommand{UploadID: "0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f", ActorUserID: "missing", ActorGroupID: "group-1"})
	if err == nil || err.Error() != "validating actor ids: [UNKNOWN_ACTOR_USER_ID] actor user id was not found" {
		t.Fatalf("Execute() error = %v, want actor validation error", err)
	}
}

func TestRetryTransactionUploadClassificationUseCaseExecuteParsesUploadID(t *testing.T) {
	t.Parallel()

	useCase := NewRetryTransactionUploadClassificationUseCase(
		&transactionUploadRepositoryMock{},
		&transactionClassificationRetryRepositoryMock{},
		&transactionManagerUseCaseMock{},
		&adminEventRecorderUploadMock{},
		&actorDirectoryMock{},
		zap.NewNop(),
	)

	_, err := useCase.Execute(context.Background(), inboundports.RetryTransactionUploadClassificationCommand{UploadID: "bad-id", ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"})
	if err == nil || err.Error() != "parsing upload id: [INVALID_UPLOAD_ID] upload id is invalid" {
		t.Fatalf("Execute() error = %v, want upload parse error", err)
	}
}

func TestRetryTransactionUploadClassificationUseCaseUsesSingleTimestampPerRequest(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transactionID1, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	transactionID2, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e61")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	retryRepo := &transactionClassificationRetryRepositoryMock{listResult: []valueobjects.TransactionID{transactionID1, transactionID2}}
	useCase := NewRetryTransactionUploadClassificationUseCase(
		&transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, valueobjects.UploadedTransactionUploadStatus(), 2)},
		retryRepo,
		&transactionManagerUseCaseMock{},
		&adminEventRecorderUploadMock{},
		&actorDirectoryMock{},
		zap.NewNop(),
	)
	fixedNow := time.Date(2026, time.April, 4, 12, 0, 0, 0, time.UTC)
	useCase.now = func() time.Time { return fixedNow }

	_, err = useCase.Execute(context.Background(), inboundports.RetryTransactionUploadClassificationCommand{UploadID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(retryRepo.retryAttempts) != 2 {
		t.Fatalf("len(retryRepo.retryAttempts) = %d, want %d", len(retryRepo.retryAttempts), 2)
	}

	for _, attempt := range retryRepo.retryAttempts {
		if !attempt.LastRetriedAt.Equal(fixedNow) {
			t.Fatalf("attempt.LastRetriedAt = %v, want %v", attempt.LastRetriedAt, fixedNow)
		}
	}
}

func TestRetryTransactionUploadClassificationUseCaseLogsSkippedDetails(t *testing.T) {
	t.Parallel()

	uploadID, err := valueobjects.UploadIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e5f")
	if err != nil {
		t.Fatalf("UploadIDFromString() error = %v", err)
	}

	transactionID, err := valueobjects.TransactionIDFromString("0195f3fd-c1a8-7366-a9f9-d81f9f2f8e60")
	if err != nil {
		t.Fatalf("TransactionIDFromString() error = %v", err)
	}

	observedCore, observedLogs := observer.New(zapcore.InfoLevel)
	useCase := NewRetryTransactionUploadClassificationUseCase(
		&transactionUploadRepositoryMock{findByIDResult: testReconstitutedTransactionUpload(t, uploadID, valueobjects.UploadedTransactionUploadStatus(), 1)},
		&transactionClassificationRetryRepositoryMock{
			listResult: []valueobjects.TransactionID{transactionID},
			retryResults: []outboundports.RetryFailedTransactionResult{{
				Skipped:    true,
				SkipReason: "already_queued_or_not_failed",
			}},
		},
		&transactionManagerUseCaseMock{},
		&adminEventRecorderUploadMock{},
		&actorDirectoryMock{},
		zap.New(observedCore),
	)
	useCase.now = testTime

	_, err = useCase.Execute(context.Background(), inboundports.RetryTransactionUploadClassificationCommand{UploadID: uploadID.String(), ActorUserID: "01962b8f-aeb2-7e03-a8ff-1edce1300002", ActorGroupID: "group-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	entries := observedLogs.All()
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want %d", len(entries), 1)
	}

	contextMap := entries[0].ContextMap()
	skippedEntries, ok := contextMap["skipped"].([]inboundports.RetryTransactionUploadClassificationSkippedTransaction)
	if !ok {
		t.Fatalf("contextMap[skipped] type = %T, want []inboundports.RetryTransactionUploadClassificationSkippedTransaction", contextMap["skipped"])
	}
	if len(skippedEntries) != 1 {
		t.Fatalf("len(skippedEntries) = %d, want %d", len(skippedEntries), 1)
	}
	if skippedEntries[0].Reason != "already_queued_or_not_failed" {
		t.Fatalf("skippedEntries[0].Reason = %q, want %q", skippedEntries[0].Reason, "already_queued_or_not_failed")
	}
}
