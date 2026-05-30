package services

import (
	"context"
	"fmt"
	"reflect"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

// TransactionUploadProgressService runs upload ingestion while reporting progress updates.
type TransactionUploadProgressService struct {
	createUpload inboundports.CreateTransactionUploadPort
}

// NewTransactionUploadProgressService builds a TransactionUploadProgressService.
func NewTransactionUploadProgressService(createUpload inboundports.CreateTransactionUploadPort) *TransactionUploadProgressService {
	return &TransactionUploadProgressService{createUpload: createUpload}
}

// Execute uploads and ingests a transaction file while reporting progress updates.
func (s *TransactionUploadProgressService) Execute(ctx context.Context, command inboundports.CreateTransactionUploadCommand, reporter outboundports.TransactionUploadProgressReporter) (inboundports.CreateTransactionUploadResult, error) {
	reportTransactionUploadProgress(ctx, reporter, outboundports.TransactionUploadProgressStatusReceived, "upload received", 5, nil, nil)
	command.ProgressReporter = mergeTransactionUploadProgressReporters(command.ProgressReporter, reporter)

	result, err := s.createUpload.Execute(ctx, command)
	if err != nil {
		return inboundports.CreateTransactionUploadResult{}, fmt.Errorf("executing transaction upload: %w", err)
	}

	return result, nil
}

type transactionUploadProgressReporterChain struct {
	reporters []outboundports.TransactionUploadProgressReporter
}

// Report fans one upload progress update out to each configured reporter.
func (c transactionUploadProgressReporterChain) Report(ctx context.Context, update outboundports.TransactionUploadProgressUpdate) error {
	// Stop on the first reporter failure so the caller can surface the transport or
	// observer problem immediately.
	for _, reporter := range c.reporters {
		if reporter == nil {
			continue
		}

		if err := reporter.Report(ctx, update); err != nil {
			return err
		}
	}

	return nil
}

func mergeTransactionUploadProgressReporters(reporters ...outboundports.TransactionUploadProgressReporter) outboundports.TransactionUploadProgressReporter {
	filtered := make([]outboundports.TransactionUploadProgressReporter, 0, len(reporters))
	for _, reporter := range reporters {
		if !isNilTransactionUploadProgressReporter(reporter) {
			filtered = append(filtered, reporter)
		}
	}

	switch len(filtered) {
	case 0:
		return nil
	case 1:
		return filtered[0]
	default:
		return transactionUploadProgressReporterChain{reporters: filtered}
	}
}

func reportTransactionUploadProgress(ctx context.Context, reporter outboundports.TransactionUploadProgressReporter, status, message string, progress int, upload *outboundports.TransactionUploadResult, validationErrors []outboundports.TransactionFileValidationError) {
	if isNilTransactionUploadProgressReporter(reporter) {
		return
	}

	_ = reporter.Report(ctx, outboundports.TransactionUploadProgressUpdate{
		Status:           status,
		Message:          message,
		Progress:         progress,
		Upload:           upload,
		ValidationErrors: validationErrors,
	})
}

func isNilTransactionUploadProgressReporter(reporter outboundports.TransactionUploadProgressReporter) bool {
	if reporter == nil {
		return true
	}

	value := reflect.ValueOf(reporter)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
