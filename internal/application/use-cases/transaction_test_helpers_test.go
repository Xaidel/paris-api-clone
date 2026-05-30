package usecases

import (
	"context"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

type transactionRepositoryMock struct {
	createdTransaction  *entities.Transaction
	createdTransactions []*entities.Transaction
	createdBy           string
	updatedTransaction  *entities.Transaction
	findByID            *entities.Transaction
	findByIDErr         error
	navigationLookup    ports.TransactionNavigationLookup
	navigationResult    *ports.TransactionNavigationResult
	navigationErr       error
	listFilter          ports.TransactionFilter
	listByUploadIDs     []valueobjects.UploadID
	listResult          []*entities.Transaction
	hasProcessingUpload valueobjects.UploadID
	hasProcessingResult bool
	hasProcessingErr    error
	createErr           error
	updateErr           error
	listErr             error
	deletedID           string
	deleteErr           error
	deletedByUploadID   string
	deleteByUploadErr   error
}

func (m *transactionRepositoryMock) Create(_ context.Context, transaction *entities.Transaction, createdByUserID string) error {
	m.createdTransaction = transaction
	m.createdBy = createdByUserID
	return m.createErr
}

func (m *transactionRepositoryMock) CreateMany(_ context.Context, transactions []*entities.Transaction, createdByUserID string) error {
	m.createdTransactions = transactions
	m.createdBy = createdByUserID
	return m.createErr
}

func (m *transactionRepositoryMock) Update(_ context.Context, transaction *entities.Transaction) error {
	m.updatedTransaction = transaction
	return m.updateErr
}

func (m *transactionRepositoryMock) FindByID(_ context.Context, _ valueobjects.TransactionID) (*entities.Transaction, error) {
	return m.findByID, m.findByIDErr
}

func (m *transactionRepositoryMock) FindHistoricalClassificationByExactGoodsDescription(context.Context, ports.HistoricalTransactionClassificationQuery) (*ports.HistoricalTransactionClassificationMatch, error) {
	return nil, nil
}

func (m *transactionRepositoryMock) GetNavigation(_ context.Context, lookup ports.TransactionNavigationLookup) (*ports.TransactionNavigationResult, error) {
	m.navigationLookup = lookup
	return m.navigationResult, m.navigationErr
}

func (m *transactionRepositoryMock) List(_ context.Context, filter ports.TransactionFilter) ([]*entities.Transaction, error) {
	m.listFilter = filter
	return m.listResult, m.listErr
}

func (m *transactionRepositoryMock) ListByUploadIDs(_ context.Context, uploadIDs []valueobjects.UploadID) ([]*entities.Transaction, error) {
	m.listByUploadIDs = uploadIDs
	return m.listResult, m.listErr
}

func (m *transactionRepositoryMock) HasProcessingByUploadID(_ context.Context, uploadID valueobjects.UploadID) (bool, error) {
	m.hasProcessingUpload = uploadID
	return m.hasProcessingResult, m.hasProcessingErr
}

func (m *transactionRepositoryMock) DeleteByID(_ context.Context, id valueobjects.TransactionID) error {
	m.deletedID = id.String()
	return m.deleteErr
}

func (m *transactionRepositoryMock) DeleteByUploadID(_ context.Context, uploadID valueobjects.UploadID) error {
	m.deletedByUploadID = uploadID.String()
	return m.deleteByUploadErr
}

func (m *transactionStep4RepositoryMock) FindByTransactionID(_ context.Context, transactionID valueobjects.TransactionID) (*entities.TransactionStep4, error) {
	_ = transactionID
	return m.findByTransactionID, m.findByTransactionIDErr
}
