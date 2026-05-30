package services

import (
	"context"
	"testing"

	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
)

type createTransactionUploadPortStub struct {
	result          inboundports.CreateTransactionUploadResult
	err             error
	receivedCommand inboundports.CreateTransactionUploadCommand
	updates         []outboundports.TransactionUploadProgressUpdate
}

func (s *createTransactionUploadPortStub) Execute(ctx context.Context, command inboundports.CreateTransactionUploadCommand) (inboundports.CreateTransactionUploadResult, error) {
	s.receivedCommand = command
	if command.ProgressReporter != nil {
		for _, update := range s.updates {
			_ = command.ProgressReporter.Report(ctx, update)
		}
	}

	return s.result, s.err
}

type transactionUploadProgressReporterStub struct {
	updates []outboundports.TransactionUploadProgressUpdate
}

func (s *transactionUploadProgressReporterStub) Report(_ context.Context, update outboundports.TransactionUploadProgressUpdate) error {
	s.updates = append(s.updates, update)
	return nil
}

// TestTransactionUploadProgressServiceExecute verifies the transaction upload progress service execute behavior and the expected outcome asserted below.
func TestTransactionUploadProgressServiceExecute(t *testing.T) {
	t.Parallel()

	createUpload := &createTransactionUploadPortStub{
		result: inboundports.CreateTransactionUploadResult{},
		updates: []outboundports.TransactionUploadProgressUpdate{{
			Status:   outboundports.TransactionUploadProgressStatusParsed,
			Message:  "file parsed",
			Progress: 25,
		}},
	}
	reporter := &transactionUploadProgressReporterStub{}
	service := NewTransactionUploadProgressService(createUpload)

	_, err := service.Execute(context.Background(), inboundports.CreateTransactionUploadCommand{}, reporter)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if createUpload.receivedCommand.ProgressReporter == nil {
		t.Fatal("received command progress reporter = nil, want reporter")
	}

	got := make([]string, 0, len(reporter.updates))
	for _, update := range reporter.updates {
		got = append(got, update.Status)
	}

	want := []string{
		outboundports.TransactionUploadProgressStatusReceived,
		outboundports.TransactionUploadProgressStatusParsed,
	}

	if len(got) != len(want) {
		t.Fatalf("len(progress statuses) = %d, want %d (%v)", len(got), len(want), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("progress status[%d] = %q, want %q (all = %v)", i, got[i], want[i], got)
		}
	}
}
