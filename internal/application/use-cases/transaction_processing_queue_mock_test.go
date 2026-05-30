package usecases

import (
	"context"

	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

type transactionProcessingQueueMock struct {
	enqueuedTasks []string
	enqueuedIDs   []string
	err           error
}

func (m *transactionProcessingQueueMock) Enqueue(_ context.Context, taskName string, transactionID valueobjects.TransactionID) error {
	m.enqueuedTasks = append(m.enqueuedTasks, taskName)
	m.enqueuedIDs = append(m.enqueuedIDs, transactionID.String())
	return m.err
}
