package usecases

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
)

type feedbackRepositoryMock struct {
	createdFeedback    *entities.Feedback
	createErr          error
	findByUserAndTx    *entities.Feedback
	findByUserAndTxErr error
	updatedFeedback    *entities.Feedback
	updateErr          error
	deletedUserID      string
	deletedTxID        string
	deleteErr          error
}

func (m *feedbackRepositoryMock) Create(_ context.Context, feedback *entities.Feedback) error {
	m.createdFeedback = feedback
	return m.createErr
}

func (m *feedbackRepositoryMock) FindByUserAndTransaction(_ context.Context, userID valueobjects.UserID, _ valueobjects.TransactionID) (*entities.Feedback, error) {
	_ = userID
	return m.findByUserAndTx, m.findByUserAndTxErr
}

func (m *feedbackRepositoryMock) FindByTransactionID(_ context.Context, _ valueobjects.TransactionID) (*entities.Feedback, error) {
	return nil, nil
}

func (m *feedbackRepositoryMock) Update(_ context.Context, feedback *entities.Feedback) error {
	m.updatedFeedback = feedback
	return m.updateErr
}

func (m *feedbackRepositoryMock) DeleteByUserAndTransaction(_ context.Context, userID valueobjects.UserID, txID valueobjects.TransactionID) error {
	m.deletedUserID = userID.String()
	m.deletedTxID = txID.String()
	return m.deleteErr
}
