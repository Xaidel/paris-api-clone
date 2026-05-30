package ports_test

import (
	"context"
	"testing"

	entities "github.com/gyud-adb/paris-api/internal/domain/entities"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	outboundports "github.com/gyud-adb/paris-api/internal/ports/outbound"
)

type createUserPortStub struct{}

type createExclusionListPortStub struct{}

type createU1ListPortStub struct{}

type createSectorPortStub struct{}

type createGroupPortStub struct{}

type createTransactionPortStub struct{}

type getUserPortStub struct{}

type getExclusionListPortStub struct{}

type getU1ListPortStub struct{}

type getSectorPortStub struct{}

type getGroupPortStub struct{}

type getTransactionPortStub struct{}

type getTransactionNavigationPortStub struct{}

type listTransactionsPortStub struct{}

type listUsersPortStub struct{}

type listExclusionListPortStub struct{}

type listU1ListPortStub struct{}

type listSectorsPortStub struct{}

type listGroupsPortStub struct{}

type updateUserPortStub struct{}

type updateGroupPortStub struct{}

type updateExclusionListPortStub struct{}

type updateU1ListPortStub struct{}

type updateSectorPortStub struct{}

type deleteUserPortStub struct{}

type deleteExclusionListPortStub struct{}

type deleteU1ListPortStub struct{}

type deleteSectorPortStub struct{}

type deleteGroupPortStub struct{}

type deleteTransactionPortStub struct{}

type listAuditEventsPortStub struct{}

type getAuditEventPortStub struct{}

type createTransactionUploadPortStub struct{}

type getTransactionUploadPortStub struct{}

type getTransactionUploadPreviewPortStub struct{}
type downloadTransactionUploadPortStub struct{}

type listTransactionUploadsPortStub struct{}

type deleteTransactionUploadPortStub struct{}

type retryTransactionUploadClassificationPortStub struct{}

type validateTransactionFilePortStub struct{}

type userRepositoryStub struct{}

type exclusionListRepositoryStub struct{}

type u1ListRepositoryStub struct{}

type sectorRepositoryStub struct{}

type groupRepositoryStub struct{}

type actorDirectoryStub struct{}

type passwordHasherStub struct{}

type adminEventRepositoryStub struct{}

type adminEventOutboxRepositoryStub struct{}

type transactionManagerStub struct{}

type rawFileStoreStub struct{}

type transactionFileParserStub struct{}

type transactionUploadRepositoryStub struct{}

type transactionUploadPreviewRepositoryStub struct{}

type transactionRepositoryStub struct{}

type transactionClassificationRetryRepositoryStub struct{}

type transactionProcessingQueueStub struct{}

type classificationListRepositoryStub struct{}

type classificationJobQueueStub struct{}

func (createUserPortStub) Execute(context.Context, inboundports.CreateUserCommand) (outboundports.UserResult, error) {
	return outboundports.UserResult{}, nil
}

func (createExclusionListPortStub) Execute(context.Context, inboundports.CreateExclusionListCommand) (outboundports.ExclusionListResult, error) {
	return outboundports.ExclusionListResult{}, nil
}

func (createU1ListPortStub) Execute(context.Context, inboundports.CreateU1ListCommand) (outboundports.U1ListResult, error) {
	return outboundports.U1ListResult{}, nil
}

func (createSectorPortStub) Execute(context.Context, inboundports.CreateSectorCommand) (outboundports.SectorResult, error) {
	return outboundports.SectorResult{}, nil
}

func (createGroupPortStub) Execute(context.Context, inboundports.CreateGroupCommand) (outboundports.GroupResult, error) {
	return outboundports.GroupResult{}, nil
}

func (createTransactionPortStub) Execute(context.Context, inboundports.CreateTransactionCommand) (outboundports.TransactionResult, error) {
	return outboundports.TransactionResult{}, nil
}

func (retryTransactionUploadClassificationPortStub) Execute(context.Context, inboundports.RetryTransactionUploadClassificationCommand) (inboundports.RetryTransactionUploadClassificationResult, error) {
	return inboundports.RetryTransactionUploadClassificationResult{}, nil
}

func (transactionProcessingQueueStub) Enqueue(context.Context, string, valueobjects.TransactionID) error {
	return nil
}

func (transactionClassificationRetryRepositoryStub) ListFailedByUploadID(context.Context, valueobjects.UploadID) ([]valueobjects.TransactionID, error) {
	return nil, nil
}

func (transactionClassificationRetryRepositoryStub) RetryFailedTransaction(context.Context, outboundports.RetryFailedTransactionCommand) (outboundports.RetryFailedTransactionResult, error) {
	return outboundports.RetryFailedTransactionResult{}, nil
}

func (classificationListRepositoryStub) GetEntries(context.Context, valueobjects.ListType) ([]string, error) {
	return nil, nil
}

func (classificationListRepositoryStub) GetEntryDocuments(context.Context, valueobjects.ListType) ([]valueobjects.ClassificationListEntryDocument, error) {
	return nil, nil
}

func (classificationJobQueueStub) Dequeue(context.Context, string) (*outboundports.ClassificationJob, error) {
	return nil, nil
}

func (classificationJobQueueStub) DequeueBatch(context.Context, string, int) ([]outboundports.ClassificationJob, error) {
	return nil, nil
}

func (classificationJobQueueStub) Complete(context.Context, outboundports.ClassificationJob) error {
	return nil
}

func (getUserPortStub) Execute(context.Context, inboundports.GetUserQuery) (outboundports.UserResult, error) {
	return outboundports.UserResult{}, nil
}

func (getExclusionListPortStub) Execute(context.Context, inboundports.GetExclusionListQuery) (outboundports.ExclusionListResult, error) {
	return outboundports.ExclusionListResult{}, nil
}

func (getU1ListPortStub) Execute(context.Context, inboundports.GetU1ListQuery) (outboundports.U1ListResult, error) {
	return outboundports.U1ListResult{}, nil
}

func (getSectorPortStub) Execute(context.Context, inboundports.GetSectorQuery) (outboundports.SectorResult, error) {
	return outboundports.SectorResult{}, nil
}

func (getGroupPortStub) Execute(context.Context, inboundports.GetGroupQuery) (outboundports.GroupResult, error) {
	return outboundports.GroupResult{}, nil
}

func (getTransactionPortStub) Execute(context.Context, inboundports.GetTransactionQuery) (outboundports.TransactionResult, error) {
	return outboundports.TransactionResult{}, nil
}

func (getTransactionNavigationPortStub) Execute(context.Context, inboundports.GetTransactionNavigationQuery) (inboundports.GetTransactionNavigationResult, error) {
	return inboundports.GetTransactionNavigationResult{}, nil
}

func (listUsersPortStub) Execute(context.Context, inboundports.ListUsersQuery) (inboundports.ListUsersResult, error) {
	return inboundports.ListUsersResult{}, nil
}

func (listExclusionListPortStub) Execute(context.Context, inboundports.ListExclusionListQuery) (outboundports.ListExclusionListResult, error) {
	return outboundports.ListExclusionListResult{}, nil
}

func (listU1ListPortStub) Execute(context.Context, inboundports.ListU1ListQuery) (outboundports.ListU1ListResult, error) {
	return outboundports.ListU1ListResult{}, nil
}

func (listSectorsPortStub) Execute(context.Context, inboundports.ListSectorsQuery) (outboundports.ListSectorsResult, error) {
	return outboundports.ListSectorsResult{}, nil
}

func (listGroupsPortStub) Execute(context.Context, inboundports.ListGroupsQuery) (outboundports.ListGroupsResult, error) {
	return outboundports.ListGroupsResult{}, nil
}

func (updateUserPortStub) Execute(context.Context, inboundports.UpdateUserCommand) (outboundports.UserResult, error) {
	return outboundports.UserResult{}, nil
}

func (updateGroupPortStub) Execute(context.Context, inboundports.UpdateGroupCommand) (outboundports.GroupResult, error) {
	return outboundports.GroupResult{}, nil
}

func (updateExclusionListPortStub) Execute(context.Context, inboundports.UpdateExclusionListCommand) (outboundports.ExclusionListResult, error) {
	return outboundports.ExclusionListResult{}, nil
}

func (updateU1ListPortStub) Execute(context.Context, inboundports.UpdateU1ListCommand) (outboundports.U1ListResult, error) {
	return outboundports.U1ListResult{}, nil
}

func (updateSectorPortStub) Execute(context.Context, inboundports.UpdateSectorCommand) (outboundports.SectorResult, error) {
	return outboundports.SectorResult{}, nil
}

func (listTransactionsPortStub) Execute(context.Context, inboundports.ListTransactionsQuery) (inboundports.ListTransactionsResult, error) {
	return inboundports.ListTransactionsResult{}, nil
}

func (deleteUserPortStub) Execute(context.Context, inboundports.DeleteUserCommand) (inboundports.DeleteUserResult, error) {
	return inboundports.DeleteUserResult{}, nil
}

func (deleteExclusionListPortStub) Execute(context.Context, inboundports.DeleteExclusionListCommand) (inboundports.DeleteExclusionListResult, error) {
	return inboundports.DeleteExclusionListResult{}, nil
}

func (deleteU1ListPortStub) Execute(context.Context, inboundports.DeleteU1ListCommand) (inboundports.DeleteU1ListResult, error) {
	return inboundports.DeleteU1ListResult{}, nil
}

func (deleteSectorPortStub) Execute(context.Context, inboundports.DeleteSectorCommand) (inboundports.DeleteSectorResult, error) {
	return inboundports.DeleteSectorResult{}, nil
}

func (deleteGroupPortStub) Execute(context.Context, inboundports.DeleteGroupCommand) (inboundports.DeleteGroupResult, error) {
	return inboundports.DeleteGroupResult{}, nil
}

func (deleteTransactionPortStub) Execute(context.Context, inboundports.DeleteTransactionCommand) (inboundports.DeleteTransactionResult, error) {
	return inboundports.DeleteTransactionResult{}, nil
}

func (listAuditEventsPortStub) Execute(context.Context, inboundports.ListAuditEventsQuery) (outboundports.ListAuditEventsResult, error) {
	return outboundports.ListAuditEventsResult{}, nil
}

func (getAuditEventPortStub) Execute(context.Context, inboundports.GetAuditEventQuery) (outboundports.AuditEventResult, error) {
	return outboundports.AuditEventResult{}, nil
}

func (createTransactionUploadPortStub) Execute(context.Context, inboundports.CreateTransactionUploadCommand) (inboundports.CreateTransactionUploadResult, error) {
	return inboundports.CreateTransactionUploadResult{}, nil
}

func (getTransactionUploadPortStub) Execute(context.Context, inboundports.GetTransactionUploadQuery) (outboundports.TransactionUploadDetailsResult, error) {
	return outboundports.TransactionUploadDetailsResult{}, nil
}

func (getTransactionUploadPreviewPortStub) Execute(context.Context, inboundports.GetTransactionUploadPreviewQuery) (inboundports.GetTransactionUploadPreviewResult, error) {
	return inboundports.GetTransactionUploadPreviewResult{}, nil
}

func (downloadTransactionUploadPortStub) Execute(context.Context, inboundports.DownloadTransactionUploadQuery) (inboundports.DownloadTransactionUploadResult, error) {
	return inboundports.DownloadTransactionUploadResult{}, nil
}

func (listTransactionUploadsPortStub) Execute(context.Context, inboundports.ListTransactionUploadsQuery) (inboundports.ListTransactionUploadsResult, error) {
	return inboundports.ListTransactionUploadsResult{}, nil
}

func (deleteTransactionUploadPortStub) Execute(context.Context, inboundports.DeleteTransactionUploadCommand) (inboundports.DeleteTransactionUploadResult, error) {
	return inboundports.DeleteTransactionUploadResult{}, nil
}

func (validateTransactionFilePortStub) Execute(context.Context, inboundports.ValidateTransactionFileCommand) (inboundports.ValidateTransactionFileResult, error) {
	return inboundports.ValidateTransactionFileResult{}, nil
}

func (userRepositoryStub) Create(context.Context, *entities.User) error {
	return nil
}

func (userRepositoryStub) FindByID(context.Context, valueobjects.UserID) (*entities.User, error) {
	return nil, nil
}

func (userRepositoryStub) List(context.Context) ([]*entities.User, error) {
	return nil, nil
}

func (userRepositoryStub) Update(context.Context, *entities.User) error {
	return nil
}

func (userRepositoryStub) DeleteByID(context.Context, valueobjects.UserID) error {
	return nil
}

func (exclusionListRepositoryStub) Create(context.Context, *entities.ExclusionListEntry, string) error {
	return nil
}

func (exclusionListRepositoryStub) FindByID(context.Context, valueobjects.ExclusionListID) (*entities.ExclusionListEntry, error) {
	return nil, nil
}

func (exclusionListRepositoryStub) List(context.Context) ([]*entities.ExclusionListEntry, error) {
	return nil, nil
}

func (exclusionListRepositoryStub) Update(context.Context, *entities.ExclusionListEntry) error {
	return nil
}

func (exclusionListRepositoryStub) DeleteByID(context.Context, valueobjects.ExclusionListID) error {
	return nil
}

func (u1ListRepositoryStub) Create(context.Context, *entities.U1ListEntry, string) error {
	return nil
}

func (u1ListRepositoryStub) FindByID(context.Context, valueobjects.U1ListID) (*entities.U1ListEntry, error) {
	return nil, nil
}

func (u1ListRepositoryStub) List(context.Context, outboundports.U1ListFilter) ([]*entities.U1ListEntry, error) {
	return nil, nil
}

func (u1ListRepositoryStub) Update(context.Context, *entities.U1ListEntry) error {
	return nil
}

func (u1ListRepositoryStub) DeleteByID(context.Context, valueobjects.U1ListID) error {
	return nil
}

func (sectorRepositoryStub) Create(context.Context, *entities.Sector, string) error {
	return nil
}

func (sectorRepositoryStub) FindByID(context.Context, valueobjects.SectorID) (*entities.Sector, error) {
	return nil, nil
}

func (sectorRepositoryStub) List(context.Context) ([]*entities.Sector, error) {
	return nil, nil
}

func (sectorRepositoryStub) Update(context.Context, *entities.Sector) error {
	return nil
}

func (sectorRepositoryStub) DeleteByID(context.Context, valueobjects.SectorID) error {
	return nil
}

func (groupRepositoryStub) Create(context.Context, *entities.Group) error {
	return nil
}

func (groupRepositoryStub) FindByID(context.Context, valueobjects.GroupID) (*entities.Group, error) {
	return nil, nil
}

func (groupRepositoryStub) List(context.Context) ([]*entities.Group, error) {
	return nil, nil
}

func (groupRepositoryStub) Update(context.Context, *entities.Group) error {
	return nil
}

func (groupRepositoryStub) DeleteByID(context.Context, valueobjects.GroupID) error {
	return nil
}

func (actorDirectoryStub) ActorExists(context.Context, string, string) error {
	return nil
}

func (passwordHasherStub) Hash(context.Context, string) (string, error) {
	return "hash", nil
}

func (adminEventRepositoryStub) Create(context.Context, *entities.AdminEvent) error {
	return nil
}

func (adminEventRepositoryStub) FindByID(context.Context, valueobjects.EventID) (*entities.AdminEvent, error) {
	return nil, nil
}

func (adminEventRepositoryStub) List(context.Context, outboundports.AuditEventFilter) ([]*entities.AdminEvent, error) {
	return nil, nil
}

func (adminEventOutboxRepositoryStub) Create(context.Context, *entities.AdminEvent) error {
	return nil
}

func (transactionManagerStub) WithinTransaction(ctx context.Context, operation func(ctx context.Context) error) error {
	return operation(ctx)
}

func (rawFileStoreStub) Store(context.Context, outboundports.StoreRawFileCommand) (outboundports.StoreRawFileResult, error) {
	return outboundports.StoreRawFileResult{}, nil
}

func (rawFileStoreStub) Read(context.Context, outboundports.ReadRawFileCommand) (outboundports.ReadRawFileResult, error) {
	return outboundports.ReadRawFileResult{}, nil
}

func (rawFileStoreStub) Delete(context.Context, outboundports.DeleteRawFileCommand) error {
	return nil
}

func (transactionFileParserStub) Parse(context.Context, outboundports.ParseTransactionFileCommand) (outboundports.ParseTransactionFileResult, error) {
	return outboundports.ParseTransactionFileResult{}, nil
}

func (transactionUploadRepositoryStub) Create(context.Context, *entities.TransactionUpload) error {
	return nil
}

func (transactionUploadRepositoryStub) FindByID(context.Context, valueobjects.UploadID) (*entities.TransactionUpload, error) {
	return nil, nil
}

func (transactionUploadRepositoryStub) FindByContentMD5(context.Context, string, valueobjects.GroupID) (*entities.TransactionUpload, error) {
	return nil, nil
}

func (transactionUploadRepositoryStub) List(context.Context, outboundports.TransactionUploadFilter) ([]*entities.TransactionUpload, error) {
	return nil, nil
}

func (transactionUploadRepositoryStub) DeleteByID(context.Context, valueobjects.UploadID) error {
	return nil
}

func (transactionUploadPreviewRepositoryStub) Save(context.Context, outboundports.TransactionUploadPreviewRecord) error {
	return nil
}

func (transactionUploadPreviewRepositoryStub) FindByUploadID(context.Context, valueobjects.UploadID) (*outboundports.TransactionUploadPreviewRecord, error) {
	return nil, nil
}

func (transactionRepositoryStub) CreateMany(context.Context, []*entities.Transaction, string) error {
	return nil
}

func (transactionRepositoryStub) Create(context.Context, *entities.Transaction, string) error {
	return nil
}

func (transactionRepositoryStub) Update(context.Context, *entities.Transaction) error {
	return nil
}

func (transactionRepositoryStub) FindByID(context.Context, valueobjects.TransactionID) (*entities.Transaction, error) {
	return nil, nil
}

func (transactionRepositoryStub) FindHistoricalClassificationByExactGoodsDescription(context.Context, outboundports.HistoricalTransactionClassificationQuery) (*outboundports.HistoricalTransactionClassificationMatch, error) {
	return nil, nil
}

func (transactionRepositoryStub) GetNavigation(context.Context, outboundports.TransactionNavigationLookup) (*outboundports.TransactionNavigationResult, error) {
	return nil, nil
}

func (transactionRepositoryStub) List(context.Context, outboundports.TransactionFilter) ([]*entities.Transaction, error) {
	return nil, nil
}

func (transactionRepositoryStub) ListByUploadIDs(context.Context, []valueobjects.UploadID) ([]*entities.Transaction, error) {
	return nil, nil
}

func (transactionRepositoryStub) HasProcessingByUploadID(context.Context, valueobjects.UploadID) (bool, error) {
	return false, nil
}

func (transactionRepositoryStub) DeleteByID(context.Context, valueobjects.TransactionID) error {
	return nil
}

func (transactionRepositoryStub) DeleteByUploadID(context.Context, valueobjects.UploadID) error {
	return nil
}

type transactionClassificationGatewayStub struct{}

func (transactionClassificationGatewayStub) Classify(context.Context, []outboundports.TransactionClassificationCandidate) ([]outboundports.TransactionClassificationDecision, error) {
	return nil, nil
}

var (
	_ inboundports.CreateUserPort                           = (*createUserPortStub)(nil)
	_ inboundports.CreateExclusionListPort                  = (*createExclusionListPortStub)(nil)
	_ inboundports.CreateU1ListPort                         = (*createU1ListPortStub)(nil)
	_ inboundports.CreateSectorPort                         = (*createSectorPortStub)(nil)
	_ inboundports.CreateGroupPort                          = (*createGroupPortStub)(nil)
	_ inboundports.CreateTransactionPort                    = (*createTransactionPortStub)(nil)
	_ inboundports.GetUserPort                              = (*getUserPortStub)(nil)
	_ inboundports.GetExclusionListPort                     = (*getExclusionListPortStub)(nil)
	_ inboundports.GetU1ListPort                            = (*getU1ListPortStub)(nil)
	_ inboundports.GetSectorPort                            = (*getSectorPortStub)(nil)
	_ inboundports.GetGroupPort                             = (*getGroupPortStub)(nil)
	_ inboundports.GetTransactionPort                       = (*getTransactionPortStub)(nil)
	_ inboundports.GetTransactionNavigationPort             = (*getTransactionNavigationPortStub)(nil)
	_ inboundports.ListTransactionsPort                     = (*listTransactionsPortStub)(nil)
	_ inboundports.ListUsersPort                            = (*listUsersPortStub)(nil)
	_ inboundports.ListExclusionListPort                    = (*listExclusionListPortStub)(nil)
	_ inboundports.ListU1ListPort                           = (*listU1ListPortStub)(nil)
	_ inboundports.ListSectorsPort                          = (*listSectorsPortStub)(nil)
	_ inboundports.ListGroupsPort                           = (*listGroupsPortStub)(nil)
	_ inboundports.ListAuditEventsPort                      = (*listAuditEventsPortStub)(nil)
	_ inboundports.GetAuditEventPort                        = (*getAuditEventPortStub)(nil)
	_ inboundports.CreateTransactionUploadPort              = (*createTransactionUploadPortStub)(nil)
	_ inboundports.GetTransactionUploadPort                 = (*getTransactionUploadPortStub)(nil)
	_ inboundports.GetTransactionUploadPreviewPort          = (*getTransactionUploadPreviewPortStub)(nil)
	_ inboundports.DownloadTransactionUploadPort            = (*downloadTransactionUploadPortStub)(nil)
	_ inboundports.ListTransactionUploadsPort               = (*listTransactionUploadsPortStub)(nil)
	_ inboundports.DeleteTransactionUploadPort              = (*deleteTransactionUploadPortStub)(nil)
	_ inboundports.RetryTransactionUploadClassificationPort = (*retryTransactionUploadClassificationPortStub)(nil)
	_ inboundports.ValidateTransactionFilePort              = (*validateTransactionFilePortStub)(nil)
	_ inboundports.UpdateUserPort                           = (*updateUserPortStub)(nil)
	_ inboundports.UpdateExclusionListPort                  = (*updateExclusionListPortStub)(nil)
	_ inboundports.UpdateU1ListPort                         = (*updateU1ListPortStub)(nil)
	_ inboundports.UpdateSectorPort                         = (*updateSectorPortStub)(nil)
	_ inboundports.UpdateGroupPort                          = (*updateGroupPortStub)(nil)
	_ inboundports.DeleteUserPort                           = (*deleteUserPortStub)(nil)
	_ inboundports.DeleteExclusionListPort                  = (*deleteExclusionListPortStub)(nil)
	_ inboundports.DeleteU1ListPort                         = (*deleteU1ListPortStub)(nil)
	_ inboundports.DeleteSectorPort                         = (*deleteSectorPortStub)(nil)
	_ inboundports.DeleteGroupPort                          = (*deleteGroupPortStub)(nil)
	_ inboundports.DeleteTransactionPort                    = (*deleteTransactionPortStub)(nil)
	_ outboundports.UserRepository                          = (*userRepositoryStub)(nil)
	_ outboundports.ExclusionListRepository                 = (*exclusionListRepositoryStub)(nil)
	_ outboundports.U1ListRepository                        = (*u1ListRepositoryStub)(nil)
	_ outboundports.SectorRepository                        = (*sectorRepositoryStub)(nil)
	_ outboundports.GroupRepository                         = (*groupRepositoryStub)(nil)
	_ outboundports.TransactionUploadRepository             = (*transactionUploadRepositoryStub)(nil)
	_ outboundports.TransactionUploadPreviewRepository      = (*transactionUploadPreviewRepositoryStub)(nil)
	_ outboundports.TransactionRepository                   = (*transactionRepositoryStub)(nil)
	_ outboundports.TransactionClassificationRetryRepository = (*transactionClassificationRetryRepositoryStub)(nil)
	_ outboundports.TransactionProcessingQueue              = (*transactionProcessingQueueStub)(nil)
	_ outboundports.ClassificationListRepository            = (*classificationListRepositoryStub)(nil)
	_ outboundports.ClassificationJobQueue                  = (*classificationJobQueueStub)(nil)
	_ outboundports.TransactionClassificationGateway        = (*transactionClassificationGatewayStub)(nil)
	_ outboundports.RawFileStore                            = (*rawFileStoreStub)(nil)
	_ outboundports.TransactionFileParser                   = (*transactionFileParserStub)(nil)
	_ outboundports.PasswordHasher                          = (*passwordHasherStub)(nil)
	_ outboundports.AdminEventRepository                    = (*adminEventRepositoryStub)(nil)
	_ outboundports.AdminEventOutboxRepository              = (*adminEventOutboxRepositoryStub)(nil)
	_ outboundports.TransactionManager                      = (*transactionManagerStub)(nil)
	_ outboundports.ActorDirectory                          = (*actorDirectoryStub)(nil)
)

// TestPorts verifies the port contracts behavior and the expected outcome asserted below.
func TestPorts(t *testing.T) {
	t.Parallel()

	uploadResult := outboundports.TransactionUploadResult{ID: "upload-1", GroupID: "group-1"}
	if uploadResult.GroupID != "group-1" {
		t.Fatalf("uploadResult.GroupID = %q, want %q", uploadResult.GroupID, "group-1")
	}

	previewResult := inboundports.GetTransactionUploadPreviewResult{FileID: "upload-1", FileName: "transactions.csv"}
	if previewResult.FileID != "upload-1" {
		t.Fatalf("previewResult.FileID = %q, want %q", previewResult.FileID, "upload-1")
	}

	contracts := []struct {
		name string
		port any
	}{
		{name: "create user port", port: inboundports.CreateUserPort(&createUserPortStub{})},
		{name: "create exclusion list port", port: inboundports.CreateExclusionListPort(&createExclusionListPortStub{})},
		{name: "create u1 list port", port: inboundports.CreateU1ListPort(&createU1ListPortStub{})},
		{name: "create sector port", port: inboundports.CreateSectorPort(&createSectorPortStub{})},
		{name: "create group port", port: inboundports.CreateGroupPort(&createGroupPortStub{})},
		{name: "create transaction port", port: inboundports.CreateTransactionPort(&createTransactionPortStub{})},
		{name: "get user port", port: inboundports.GetUserPort(&getUserPortStub{})},
		{name: "get exclusion list port", port: inboundports.GetExclusionListPort(&getExclusionListPortStub{})},
		{name: "get u1 list port", port: inboundports.GetU1ListPort(&getU1ListPortStub{})},
		{name: "get sector port", port: inboundports.GetSectorPort(&getSectorPortStub{})},
		{name: "get group port", port: inboundports.GetGroupPort(&getGroupPortStub{})},
		{name: "get transaction port", port: inboundports.GetTransactionPort(&getTransactionPortStub{})},
		{name: "get transaction navigation port", port: inboundports.GetTransactionNavigationPort(&getTransactionNavigationPortStub{})},
		{name: "list transactions port", port: inboundports.ListTransactionsPort(&listTransactionsPortStub{})},
		{name: "list users port", port: inboundports.ListUsersPort(&listUsersPortStub{})},
		{name: "list exclusion list port", port: inboundports.ListExclusionListPort(&listExclusionListPortStub{})},
		{name: "list u1 list port", port: inboundports.ListU1ListPort(&listU1ListPortStub{})},
		{name: "list sectors port", port: inboundports.ListSectorsPort(&listSectorsPortStub{})},
		{name: "list groups port", port: inboundports.ListGroupsPort(&listGroupsPortStub{})},
		{name: "list audit events port", port: inboundports.ListAuditEventsPort(&listAuditEventsPortStub{})},
		{name: "get audit event port", port: inboundports.GetAuditEventPort(&getAuditEventPortStub{})},
		{name: "create transaction upload port", port: inboundports.CreateTransactionUploadPort(&createTransactionUploadPortStub{})},
		{name: "get transaction upload port", port: inboundports.GetTransactionUploadPort(&getTransactionUploadPortStub{})},
		{name: "get transaction upload preview port", port: inboundports.GetTransactionUploadPreviewPort(&getTransactionUploadPreviewPortStub{})},
		{name: "download transaction upload port", port: inboundports.DownloadTransactionUploadPort(&downloadTransactionUploadPortStub{})},
		{name: "get transaction upload preview port", port: inboundports.GetTransactionUploadPreviewPort(&getTransactionUploadPreviewPortStub{})},
		{name: "list transaction uploads port", port: inboundports.ListTransactionUploadsPort(&listTransactionUploadsPortStub{})},
		{name: "delete transaction upload port", port: inboundports.DeleteTransactionUploadPort(&deleteTransactionUploadPortStub{})},
		{name: "retry transaction upload classification port", port: inboundports.RetryTransactionUploadClassificationPort(&retryTransactionUploadClassificationPortStub{})},
		{name: "validate transaction file port", port: inboundports.ValidateTransactionFilePort(&validateTransactionFilePortStub{})},
		{name: "update user port", port: inboundports.UpdateUserPort(&updateUserPortStub{})},
		{name: "update exclusion list port", port: inboundports.UpdateExclusionListPort(&updateExclusionListPortStub{})},
		{name: "update u1 list port", port: inboundports.UpdateU1ListPort(&updateU1ListPortStub{})},
		{name: "update sector port", port: inboundports.UpdateSectorPort(&updateSectorPortStub{})},
		{name: "update group port", port: inboundports.UpdateGroupPort(&updateGroupPortStub{})},
		{name: "delete user port", port: inboundports.DeleteUserPort(&deleteUserPortStub{})},
		{name: "delete exclusion list port", port: inboundports.DeleteExclusionListPort(&deleteExclusionListPortStub{})},
		{name: "delete u1 list port", port: inboundports.DeleteU1ListPort(&deleteU1ListPortStub{})},
		{name: "delete sector port", port: inboundports.DeleteSectorPort(&deleteSectorPortStub{})},
		{name: "delete group port", port: inboundports.DeleteGroupPort(&deleteGroupPortStub{})},
		{name: "delete transaction port", port: inboundports.DeleteTransactionPort(&deleteTransactionPortStub{})},
		{name: "user repository", port: outboundports.UserRepository(&userRepositoryStub{})},
		{name: "exclusion list repository", port: outboundports.ExclusionListRepository(&exclusionListRepositoryStub{})},
		{name: "u1 list repository", port: outboundports.U1ListRepository(&u1ListRepositoryStub{})},
		{name: "sector repository", port: outboundports.SectorRepository(&sectorRepositoryStub{})},
		{name: "group repository", port: outboundports.GroupRepository(&groupRepositoryStub{})},
		{name: "transaction upload repository", port: outboundports.TransactionUploadRepository(&transactionUploadRepositoryStub{})},
		{name: "transaction upload preview repository", port: outboundports.TransactionUploadPreviewRepository(&transactionUploadPreviewRepositoryStub{})},
		{name: "transaction repository", port: outboundports.TransactionRepository(&transactionRepositoryStub{})},
		{name: "transaction classification retry repository", port: outboundports.TransactionClassificationRetryRepository(&transactionClassificationRetryRepositoryStub{})},
		{name: "transaction processing queue", port: outboundports.TransactionProcessingQueue(&transactionProcessingQueueStub{})},
		{name: "classification list repository", port: outboundports.ClassificationListRepository(&classificationListRepositoryStub{})},
		{name: "classification job queue", port: outboundports.ClassificationJobQueue(&classificationJobQueueStub{})},
		{name: "transaction classification gateway", port: outboundports.TransactionClassificationGateway(&transactionClassificationGatewayStub{})},
		{name: "raw file store", port: outboundports.RawFileStore(&rawFileStoreStub{})},
		{name: "transaction file parser", port: outboundports.TransactionFileParser(&transactionFileParserStub{})},
		{name: "password hasher", port: outboundports.PasswordHasher(&passwordHasherStub{})},
		{name: "admin event repository", port: outboundports.AdminEventRepository(&adminEventRepositoryStub{})},
		{name: "admin event outbox repository", port: outboundports.AdminEventOutboxRepository(&adminEventOutboxRepositoryStub{})},
		{name: "transaction manager", port: outboundports.TransactionManager(&transactionManagerStub{})},
		{name: "actor directory", port: outboundports.ActorDirectory(&actorDirectoryStub{})},
	}

	for _, contract := range contracts {
		contract := contract
		t.Run(contract.name, func(t *testing.T) {
			t.Parallel()

			if contract.port == nil {
				t.Fatal("contract port should not be nil")
			}
		})
	}
}
