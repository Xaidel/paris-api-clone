package di

import (
	"context"
	"fmt"
	"net/http"

	openaimodel "github.com/cloudwego/eino-ext/components/model/openai"
	outboundadapters "github.com/gyud-adb/paris-api/internal/adapters/outbound"
	inboundadapters "github.com/gyud-adb/paris-api/internal/adapters/inbound"
	services "github.com/gyud-adb/paris-api/internal/application/services"
	usecases "github.com/gyud-adb/paris-api/internal/application/use-cases"
	domainservices "github.com/gyud-adb/paris-api/internal/domain/services"
	valueobjects "github.com/gyud-adb/paris-api/internal/domain/value-objects"
	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
	infraDB "github.com/gyud-adb/paris-api/internal/infrastructure/db"
	"github.com/gyud-adb/paris-api/internal/infrastructure/httpserver"
	"github.com/gyud-adb/paris-api/internal/infrastructure/observability"
	ports "github.com/gyud-adb/paris-api/internal/ports/outbound"
	inboundports "github.com/gyud-adb/paris-api/internal/ports/inbound"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

var newOpenAIReActChatModel = outboundadapters.NewOpenAIReActChatModel

// Application holds the fully wired runtime dependencies.
type Application struct {
	Config                       *config.AppConfig
	Logger                       *zap.Logger
	cleanupLogger                func() error
	Pool                         *pgxpool.Pool
	Router                       http.Handler
	Server                       *http.Server
	ReActClassificationWorker    *services.ReActClassificationWorker
	ClassificationListRepository ports.ClassificationListRepository
	ValidateTransactionFile      inboundports.ValidateTransactionFilePort
	CreateTransaction            inboundports.CreateTransactionPort
	CreateTransactionStep4       inboundports.CreateTransactionStep4Port
	CreateTransactionStep5       inboundports.CreateTransactionStep5Port
	GetTransaction               inboundports.GetTransactionPort
	ListTransactions             inboundports.ListTransactionsPort
	DeleteTransaction            inboundports.DeleteTransactionPort
	CreateTransactionUpload      inboundports.CreateTransactionUploadPort
	GetTransactionUpload         inboundports.GetTransactionUploadPort
	GetTransactionUploadPreview  inboundports.GetTransactionUploadPreviewPort
	ListTransactionUploads       inboundports.ListTransactionUploadsPort
	DownloadTransactionUpload    inboundports.DownloadTransactionUploadPort
	DeleteTransactionUpload      inboundports.DeleteTransactionUploadPort
	CreateU1ListEntry            inboundports.CreateU1ListPort
	GetU1ListEntry               inboundports.GetU1ListPort
	ListU1ListEntries            inboundports.ListU1ListPort
	UpdateU1ListEntry            inboundports.UpdateU1ListPort
	DeleteU1ListEntry            inboundports.DeleteU1ListPort
	CreateGroupEntry             inboundports.CreateGroupPort
	GetGroupEntry                inboundports.GetGroupPort
	ListGroupEntries             inboundports.ListGroupsPort
	UpdateGroupEntry             inboundports.UpdateGroupPort
	DeleteGroupEntry             inboundports.DeleteGroupPort
	CreateSectorEntry            inboundports.CreateSectorPort
	GetSectorEntry               inboundports.GetSectorPort
	ListSectorEntries            inboundports.ListSectorsPort
	UpdateSectorEntry            inboundports.UpdateSectorPort
	DeleteSectorEntry            inboundports.DeleteSectorPort
	CreateExclusionListEntry     inboundports.CreateExclusionListPort
	GetExclusionListEntry        inboundports.GetExclusionListPort
	ListExclusionListEntries     inboundports.ListExclusionListPort
	UpdateExclusionListEntry     inboundports.UpdateExclusionListPort
	DeleteExclusionListEntry     inboundports.DeleteExclusionListPort
	CreateBugReport              inboundports.CreateBugReportPort
	GetBugReport                 inboundports.GetBugReportPort
	ListBugReports               inboundports.ListBugReportsPort
	UpdateBugReport              inboundports.UpdateBugReportPort
	DeleteBugReport              inboundports.DeleteBugReportPort
	UpsertTransactionFeedback    inboundports.UpsertTransactionFeedbackPort
	DeleteTransactionFeedback    inboundports.DeleteTransactionFeedbackPort
	GetTransactionFeedback       inboundports.GetTransactionFeedbackPort
}

// Bootstrap loads infrastructure dependencies and wires the application graph.
func Bootstrap(ctx context.Context) (*Application, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger, cleanupLogger, err := observability.NewLogger(cfg.Log)
	if err != nil {
		return nil, err
	}
	// Build shared infrastructure first so every downstream constructor receives
	// the same logger, metrics, and database primitives.
	classificationMetrics := observability.NewClassificationMetrics()

	migrator := infraDB.NewMigrator(cfg.Database)
	if err := migrator.Up(); err != nil {
		_ = cleanupLogger()
		return nil, err
	}

	pool, err := infraDB.NewPool(ctx, cfg.Database)
	if err != nil {
		_ = cleanupLogger()
		return nil, err
	}

	// Construct outbound adapters up front. The rest of the wiring graph reuses
	// these concrete dependencies through ports.
	userRepository := outboundadapters.NewPostgresUserRepository(pool)
	groupRepository := outboundadapters.NewPostgresGroupRepository(pool)
	adminEventRepository := outboundadapters.NewPostgresAdminEventRepository(pool)
	adminEventOutboxRepository := outboundadapters.NewPostgresAdminEventOutboxRepository(pool)
	transactionUploadRepository := outboundadapters.NewPostgresTransactionUploadRepository(pool)
	transactionRepository := outboundadapters.NewPostgresTransactionRepository(pool)
	transactionClassificationRetryRepository := outboundadapters.NewPostgresTransactionClassificationRetryRepository(pool)
	transactionProcessingQueue := outboundadapters.NewPostgresTransactionProcessingQueue(pool)
	classificationJobQueue := outboundadapters.NewPostgresClassificationJobQueue(pool)
	transactionStep4Repository := outboundadapters.NewPostgresTransactionStep4Repository(pool)
	transactionStep5Repository := outboundadapters.NewPostgresTransactionStep5Repository(pool)
	u1ListRepository := outboundadapters.NewPostgresU1ListRepository(pool)
	sectorRepository := outboundadapters.NewPostgresSectorRepository(pool)
	exclusionListRepository := outboundadapters.NewPostgresExclusionListRepository(pool)
	reportRepository := outboundadapters.NewPostgresBugReportRepository(pool)
	feedbackRepository := outboundadapters.NewPostgresFeedbackRepository(pool)
	classificationListRepository := outboundadapters.NewPostgresClassificationListRepository(pool)
	transactionUploadPreviewRepository := outboundadapters.NewPostgresTransactionUploadPreviewRepository(pool)
	passwordHasher := outboundadapters.NewBcryptPasswordHasher()
	actorDirectory := outboundadapters.NewPostgresActorDirectory(pool)
	transactionManager := outboundadapters.NewPgxTransactionManager(pool)
	eventRecorderService := services.NewEventRecorderService(adminEventRepository, adminEventOutboxRepository, actorDirectory)
	transactionAuditService := services.NewTransactionAuditService(eventRecorderService)
	transactionFileValidator := domainservices.NewTransactionFileValidator(valueobjects.TransactionFileSchemaV1())
	transactionFileParser := outboundadapters.NewTransactionFileParser()
	rawFileStore, err := buildRawFileStore(ctx, cfg.Storage)
	if err != nil {
		pool.Close()
		_ = cleanupLogger()
		return nil, err
	}
	var _ ports.ClassificationListRepository = classificationListRepository
	reactChatModel, err := buildReActChatModel(ctx, cfg.Classification)
	if err != nil {
		logger.Warn("react chat model unavailable; react classification worker will run in degraded mode", zap.Error(err))
	}
	reactSystemPromptBuilder := outboundadapters.NewReActTransactionClassificationSystemPromptBuilder(
		cfg.Classification.ReactSystemPrompt,
		u1ListRepository,
		exclusionListRepository,
	)
	reactClassificationGateway := outboundadapters.NewReActTransactionClassificationGateway(
		transactionRepository,
		reactChatModel,
		outboundadapters.WithReActTransactionClassificationSystemPromptBuilder(reactSystemPromptBuilder),
		outboundadapters.WithReActTransactionClassificationBatchSize(cfg.Classification.ReactBatchSize),
		outboundadapters.WithReActTransactionClassificationClassifier(cfg.Classification.ReactClassifierFamily, cfg.Classification.ReactClassifierVersion),
		outboundadapters.WithReActTransactionClassificationPromptVersion(cfg.Classification.ReactPromptVersion),
		outboundadapters.WithReActTransactionClassificationModelName(cfg.Classification.ReactModel),
		outboundadapters.WithReActTransactionClassificationRetry(cfg.Classification.ReactMaxRetries, cfg.Classification.ReactRetryBackoff),
		outboundadapters.WithReActTransactionClassificationMetrics(classificationMetrics),
		outboundadapters.WithReActTransactionClassificationLogger(logger),
	)
	reactClassificationJobHandler := services.NewReActClassificationJobHandler(reactClassificationGateway, transactionRepository, logger)
	reactClassificationWorker := services.NewReActClassificationWorker(
		classificationJobQueue,
		reactClassificationJobHandler,
		transactionManager,
		logger,
		services.WithReActClassificationWorkerBatchSize(cfg.Classification.ReactBatchSize),
		services.WithReActClassificationWorkerFlushTimeout(cfg.Classification.ReactFlushTimeout),
	)
	// Expose application capabilities as use cases after the shared services and
	// repositories are ready.
	createUserUseCase := usecases.NewCreateUserUseCase(userRepository, groupRepository, passwordHasher, eventRecorderService, transactionManager, actorDirectory)
	getUserUseCase := usecases.NewGetUserUseCase(userRepository, eventRecorderService, actorDirectory)
	listUsersUseCase := usecases.NewListUsersUseCase(userRepository, eventRecorderService, actorDirectory)
	updateUserUseCase := usecases.NewUpdateUserUseCase(userRepository, groupRepository, passwordHasher, transactionManager, eventRecorderService, actorDirectory)
	deleteUserUseCase := usecases.NewDeleteUserUseCase(userRepository, eventRecorderService, actorDirectory)
	createGroupUseCase := usecases.NewCreateGroupUseCase(groupRepository, transactionManager, eventRecorderService, actorDirectory)
	getGroupUseCase := usecases.NewGetGroupUseCase(groupRepository, eventRecorderService, actorDirectory)
	listGroupsUseCase := usecases.NewListGroupsUseCase(groupRepository, eventRecorderService, actorDirectory)
	updateGroupUseCase := usecases.NewUpdateGroupUseCase(groupRepository, transactionManager, eventRecorderService, actorDirectory)
	deleteGroupUseCase := usecases.NewDeleteGroupUseCase(groupRepository, transactionManager, eventRecorderService, actorDirectory)
	createU1ListUseCase := usecases.NewCreateU1ListUseCase(u1ListRepository, transactionManager, eventRecorderService, actorDirectory)
	getU1ListUseCase := usecases.NewGetU1ListUseCase(u1ListRepository, eventRecorderService)
	listU1ListUseCase := usecases.NewListU1ListUseCase(u1ListRepository, eventRecorderService)
	updateU1ListUseCase := usecases.NewUpdateU1ListUseCase(u1ListRepository, transactionManager, eventRecorderService)
	deleteU1ListUseCase := usecases.NewDeleteU1ListUseCase(u1ListRepository, transactionManager, eventRecorderService)
	createSectorUseCase := usecases.NewCreateSectorUseCase(sectorRepository, transactionManager, eventRecorderService, actorDirectory)
	getSectorUseCase := usecases.NewGetSectorUseCase(sectorRepository, eventRecorderService)
	listSectorsUseCase := usecases.NewListSectorsUseCase(sectorRepository, eventRecorderService)
	updateSectorUseCase := usecases.NewUpdateSectorUseCase(sectorRepository, transactionManager, eventRecorderService)
	deleteSectorUseCase := usecases.NewDeleteSectorUseCase(sectorRepository, transactionManager, eventRecorderService)
	createExclusionListUseCase := usecases.NewCreateExclusionListUseCase(exclusionListRepository, transactionManager, eventRecorderService, actorDirectory)
	getExclusionListUseCase := usecases.NewGetExclusionListUseCase(exclusionListRepository, eventRecorderService)
	listExclusionListUseCase := usecases.NewListExclusionListUseCase(exclusionListRepository, eventRecorderService)
	updateExclusionListUseCase := usecases.NewUpdateExclusionListUseCase(exclusionListRepository, eventRecorderService)
	deleteExclusionListUseCase := usecases.NewDeleteExclusionListUseCase(exclusionListRepository, eventRecorderService)
	createBugReportUseCase := usecases.NewCreateBugReportUseCase(reportRepository, transactionRepository, transactionManager, eventRecorderService)
	getBugReportUseCase := usecases.NewGetBugReportUseCase(reportRepository, eventRecorderService)
	listBugReportsUseCase := usecases.NewListBugReportsUseCase(reportRepository, eventRecorderService)
	updateBugReportUseCase := usecases.NewUpdateBugReportUseCase(reportRepository, transactionManager, eventRecorderService)
	deleteBugReportUseCase := usecases.NewDeleteBugReportUseCase(reportRepository, transactionManager, eventRecorderService)
	upsertTransactionFeedbackUseCase := usecases.NewUpsertTransactionFeedbackUseCase(feedbackRepository, transactionRepository, transactionManager, eventRecorderService)
	deleteTransactionFeedbackUseCase := usecases.NewDeleteTransactionFeedbackUseCase(feedbackRepository, transactionManager, eventRecorderService)
	getTransactionFeedbackUseCase := usecases.NewGetTransactionFeedbackUseCase(feedbackRepository, eventRecorderService)
	getAuditEventUseCase := usecases.NewGetAuditEventUseCase(adminEventRepository)
	listAuditEventsUseCase := usecases.NewListAuditEventsUseCase(adminEventRepository)
	validateTransactionFileUseCase := usecases.NewValidateTransactionFileUseCase(transactionFileValidator)
	createTransactionUseCase := usecases.NewCreateTransactionUseCase(transactionRepository, transactionProcessingQueue, transactionManager, eventRecorderService, actorDirectory, transactionFileValidator)
	createTransactionStep4UseCase := usecases.NewCreateTransactionStep4UseCase(transactionStep4Repository, sectorRepository, transactionRepository, transactionManager, eventRecorderService)
	createTransactionStep5UseCase := usecases.NewCreateTransactionStep5UseCase(transactionStep5Repository, transactionStep4Repository, transactionRepository, transactionManager, eventRecorderService)
	getTransactionUseCase := usecases.NewGetTransactionUseCase(transactionRepository, transactionStep4Repository, transactionStep5Repository, sectorRepository, eventRecorderService)
	getTransactionNavigationUseCase := usecases.NewGetTransactionNavigationUseCase(transactionRepository)
	listTransactionsUseCase := usecases.NewListTransactionsUseCase(transactionRepository, transactionStep4Repository, transactionStep5Repository, sectorRepository, transactionAuditService)
	deleteTransactionUseCase := usecases.NewDeleteTransactionUseCase(transactionRepository, transactionManager, eventRecorderService)
	createTransactionUploadUseCase := usecases.NewCreateTransactionUploadUseCase(transactionUploadRepository, transactionUploadPreviewRepository, transactionRepository, transactionProcessingQueue, rawFileStore, transactionFileParser, transactionManager, eventRecorderService, actorDirectory, transactionFileValidator)
	transactionUploadProgressService := services.NewTransactionUploadProgressService(createTransactionUploadUseCase)
	getTransactionUploadUseCase := usecases.NewGetTransactionUploadUseCase(transactionUploadRepository, transactionRepository, transactionStep4Repository, transactionStep5Repository, sectorRepository, eventRecorderService, actorDirectory)
	getTransactionUploadPreviewUseCase := usecases.NewGetTransactionUploadPreviewUseCase(transactionUploadRepository, transactionUploadPreviewRepository, eventRecorderService, actorDirectory)
	listTransactionUploadsUseCase := usecases.NewListTransactionUploadsUseCase(transactionUploadRepository, transactionRepository, transactionStep4Repository, transactionStep5Repository, sectorRepository)
	downloadTransactionUploadUseCase := usecases.NewDownloadTransactionUploadUseCase(transactionUploadRepository, rawFileStore, eventRecorderService, actorDirectory)
	deleteTransactionUploadUseCase := usecases.NewDeleteTransactionUploadUseCase(transactionUploadRepository, transactionRepository, rawFileStore, transactionManager, eventRecorderService, actorDirectory)
	retryTransactionUploadClassificationUseCase := usecases.NewRetryTransactionUploadClassificationUseCase(transactionUploadRepository, transactionClassificationRetryRepository, transactionManager, eventRecorderService, actorDirectory, logger)
	httpTransactionAdapter := inboundadapters.NewHttpTransactionAdapter(createTransactionUseCase, getTransactionUseCase, getTransactionNavigationUseCase, listTransactionsUseCase, deleteTransactionUseCase)
	httpTransactionStep4Adapter := inboundadapters.NewHttpTransactionStep4Adapter(createTransactionStep4UseCase)
	httpTransactionStep5Adapter := inboundadapters.NewHttpTransactionStep5Adapter(createTransactionStep5UseCase)
	httpUserAdapter := inboundadapters.NewHttpUserAdapter(createUserUseCase, getUserUseCase, listUsersUseCase, updateUserUseCase, deleteUserUseCase)
	httpGroupAdapter := inboundadapters.NewHttpGroupAdapter(createGroupUseCase, getGroupUseCase, listGroupsUseCase, updateGroupUseCase, deleteGroupUseCase)
	httpU1ListAdapter := inboundadapters.NewHttpU1ListAdapter(createU1ListUseCase, getU1ListUseCase, listU1ListUseCase, updateU1ListUseCase, deleteU1ListUseCase)
	httpSectorAdapter := inboundadapters.NewHttpSectorAdapter(createSectorUseCase, getSectorUseCase, listSectorsUseCase, updateSectorUseCase, deleteSectorUseCase)
	httpExclusionListAdapter := inboundadapters.NewHttpExclusionListAdapter(createExclusionListUseCase, getExclusionListUseCase, listExclusionListUseCase, updateExclusionListUseCase, deleteExclusionListUseCase)
	httpBugReportAdapter := inboundadapters.NewHttpBugReportAdapter(createBugReportUseCase, getBugReportUseCase, listBugReportsUseCase, updateBugReportUseCase, deleteBugReportUseCase)
	httpTransactionFeedbackAdapter := inboundadapters.NewHttpTransactionFeedbackAdapter(upsertTransactionFeedbackUseCase, deleteTransactionFeedbackUseCase, getTransactionFeedbackUseCase)
	httpAuditEventAdapter := inboundadapters.NewHttpAuditEventAdapter(listAuditEventsUseCase, getAuditEventUseCase)
	httpTransactionUploadAdapter := inboundadapters.NewHttpTransactionUploadAdapter(createTransactionUploadUseCase, transactionUploadProgressService, getTransactionUploadUseCase, getTransactionUploadPreviewUseCase, listTransactionUploadsUseCase, deleteTransactionUploadUseCase, retryTransactionUploadClassificationUseCase, downloadTransactionUploadUseCase)

	router, err := httpserver.NewRouter(logger, httpUserAdapter, httpGroupAdapter, httpU1ListAdapter, httpSectorAdapter, httpExclusionListAdapter, httpAuditEventAdapter, httpTransactionAdapter, httpTransactionStep4Adapter, httpTransactionStep5Adapter, httpTransactionUploadAdapter, httpBugReportAdapter, httpTransactionFeedbackAdapter)
	if err != nil {
		pool.Close()
		_ = cleanupLogger()
		return nil, err
	}

	server := NewHTTPServer(serverAddress(cfg.HTTP.Port), cfg.HTTP, router)
	return &Application{
		Config:                       cfg,
		Logger:                       logger,
		cleanupLogger:                cleanupLogger,
		Pool:                         pool,
		Router:                       router,
		Server:                       server,
		ReActClassificationWorker:    reactClassificationWorker,
		ClassificationListRepository: classificationListRepository,
		ValidateTransactionFile:      validateTransactionFileUseCase,
		CreateTransaction:            createTransactionUseCase,
		CreateTransactionStep4:       createTransactionStep4UseCase,
		CreateTransactionStep5:       createTransactionStep5UseCase,
		GetTransaction:               getTransactionUseCase,
		ListTransactions:             listTransactionsUseCase,
		DeleteTransaction:            deleteTransactionUseCase,
		CreateTransactionUpload:      createTransactionUploadUseCase,
		GetTransactionUpload:         getTransactionUploadUseCase,
		GetTransactionUploadPreview:  getTransactionUploadPreviewUseCase,
		ListTransactionUploads:       listTransactionUploadsUseCase,
		DownloadTransactionUpload:    downloadTransactionUploadUseCase,
		DeleteTransactionUpload:      deleteTransactionUploadUseCase,
		CreateU1ListEntry:            createU1ListUseCase,
		GetU1ListEntry:               getU1ListUseCase,
		ListU1ListEntries:            listU1ListUseCase,
		UpdateU1ListEntry:            updateU1ListUseCase,
		DeleteU1ListEntry:            deleteU1ListUseCase,
		CreateGroupEntry:             createGroupUseCase,
		GetGroupEntry:                getGroupUseCase,
		ListGroupEntries:             listGroupsUseCase,
		UpdateGroupEntry:             updateGroupUseCase,
		DeleteGroupEntry:             deleteGroupUseCase,
		CreateSectorEntry:            createSectorUseCase,
		GetSectorEntry:               getSectorUseCase,
		ListSectorEntries:            listSectorsUseCase,
		UpdateSectorEntry:            updateSectorUseCase,
		DeleteSectorEntry:            deleteSectorUseCase,
		CreateExclusionListEntry:     createExclusionListUseCase,
		GetExclusionListEntry:        getExclusionListUseCase,
		ListExclusionListEntries:     listExclusionListUseCase,
		UpdateExclusionListEntry:     updateExclusionListUseCase,
		DeleteExclusionListEntry:     deleteExclusionListUseCase,
		CreateBugReport:              createBugReportUseCase,
		GetBugReport:                 getBugReportUseCase,
		ListBugReports:               listBugReportsUseCase,
		UpdateBugReport:              updateBugReportUseCase,
		DeleteBugReport:              deleteBugReportUseCase,
		UpsertTransactionFeedback:    upsertTransactionFeedbackUseCase,
		DeleteTransactionFeedback:    deleteTransactionFeedbackUseCase,
		GetTransactionFeedback:       getTransactionFeedbackUseCase,
	}, nil
}

func buildRawFileStore(ctx context.Context, cfg config.StorageConfig) (ports.RawFileStore, error) {
	_ = ctx

	// Keep provider selection here so the rest of the application only depends on
	// the RawFileStore port.
	switch cfg.Provider {
	case "local":
		return outboundadapters.NewLocalRawFileStore(cfg.LocalTransactionPath), nil
	case "azure_blob":
		store, err := outboundadapters.NewAzureBlobRawFileStore(cfg.AzureBlobConnection, cfg.AzureBlobContainer)
		if err != nil {
			return nil, fmt.Errorf("creating azure blob raw file store: %w", err)
		}

		return store, nil
	default:
		return nil, fmt.Errorf("unsupported raw file storage provider %q", cfg.Provider)
	}
}

func reActChatModelConfig(cfg config.ClassificationConfig) *openaimodel.ChatModelConfig {
	return &openaimodel.ChatModelConfig{
		APIKey:     cfg.OpenAIAPIKey,
		ByAzure:    cfg.OpenAIUseAzure,
		BaseURL:    cfg.OpenAIBaseURL,
		APIVersion: cfg.OpenAIAPIVersion,
		Model:      cfg.ReactModel,
		Timeout:    cfg.ReactRequestTimeout,
	}
}

func buildReActChatModel(ctx context.Context, cfg config.ClassificationConfig) (*openaimodel.ChatModel, error) {
	return newOpenAIReActChatModel(ctx, reActChatModelConfig(cfg))
}

// Shutdown releases infrastructure dependencies.
func (a *Application) Shutdown() error {
	if a == nil {
		return nil
	}

	var cleanupErr error

	if a.Pool != nil {
		a.Pool.Close()
	}

	if a.cleanupLogger != nil {
		if err := a.cleanupLogger(); err != nil {
			cleanupErr = fmt.Errorf("cleaning up logger: %w", err)
		}
	}

	return cleanupErr
}

// NewHTTPServer builds the HTTP server.
func NewHTTPServer(address string, cfg config.HTTPConfig, handler http.Handler) *http.Server {
	if handler == nil {
		// Tests and bootstrap helpers can omit a handler when they only need server
		// configuration; use an empty mux to keep construction total.
		handler = http.NewServeMux()
	}

	return &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

// serverAddress formats the server address from a port.
func serverAddress(port string) string {
	return ":" + port
}
