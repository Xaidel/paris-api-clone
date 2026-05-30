package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// AppConfig groups all runtime configuration required to bootstrap the service.
type AppConfig struct {
	Env            string
	ServiceName    string
	HTTP           HTTPConfig
	Database       DatabaseConfig
	Classification ClassificationConfig
	Storage        StorageConfig
	Log            LogConfig
}

// HTTPConfig controls listener and timeout behavior for the public HTTP server.
type HTTPConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds database connectivity and migration settings.
type DatabaseConfig struct {
	URL            string
	PingTimeout    time.Duration
	MigrationsPath string
}

// ClassificationConfig tunes the transaction classification pipeline and ReAct
// worker behavior.
type ClassificationConfig struct {
	OpenAIAPIKey           string
	OpenAIBaseURL          string
	OpenAIAPIVersion       string
	OpenAIUseAzure         bool
	ReactModel             string
	ReactSystemPrompt      string
	ReactBatchSize         int
	ReactFlushTimeout      time.Duration
	ReactClassifierFamily  string
	ReactClassifierVersion string
	ReactPromptVersion     string
	ReactRequestTimeout    time.Duration
	ReactMaxRetries        int
	ReactRetryBackoff      time.Duration
}

// StorageConfig selects and configures the raw uploaded-file backend.
type StorageConfig struct {
	Provider             string
	LocalTransactionPath string
	AzureBlobConnection  string
	AzureBlobContainer   string
}

// LogConfig configures structured logging output and metadata.
type LogConfig struct {
	Level       string
	FilePath    string
	ServiceName string
	Environment string
}

// Load loads application configuration from the environment.
func Load() (*AppConfig, error) {
	env := strings.TrimSpace(os.Getenv("ENV"))
	if env == "" || env == "development" {
		// Local development is the only mode that opportunistically loads .env.
		if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading development environment file: %w", err)
		}
	}

	env = strings.TrimSpace(os.Getenv("ENV"))
	if env == "" {
		return nil, fmt.Errorf("ENV is required")
	}

	serviceName, err := requiredEnv("SERVICE_NAME")
	if err != nil {
		return nil, err
	}

	port, err := requiredEnv("HTTP_PORT")
	if err != nil {
		return nil, err
	}

	readTimeout, err := durationEnv("HTTP_READ_TIMEOUT")
	if err != nil {
		return nil, err
	}

	writeTimeout, err := durationEnv("HTTP_WRITE_TIMEOUT")
	if err != nil {
		return nil, err
	}

	idleTimeout, err := durationEnv("HTTP_IDLE_TIMEOUT")
	if err != nil {
		return nil, err
	}

	shutdownTimeout, err := durationEnv("HTTP_SHUTDOWN_TIMEOUT")
	if err != nil {
		return nil, err
	}

	databaseURL, err := requiredEnv("DATABASE_URL")
	if err != nil {
		return nil, err
	}

	pingTimeout, err := durationEnv("DATABASE_PING_TIMEOUT")
	if err != nil {
		return nil, err
	}

	migrationsPath, err := requiredEnv("MIGRATIONS_PATH")
	if err != nil {
		return nil, err
	}

	logLevel, err := requiredEnv("LOG_LEVEL")
	if err != nil {
		return nil, err
	}

	logFilePath, err := requiredEnv("LOG_FILE_PATH")
	if err != nil {
		return nil, err
	}

	storageProvider, err := requiredEnv("STORAGE_PROVIDER")
	if err != nil {
		return nil, err
	}

	localTransactionPath, err := requiredEnv("LOCAL_TRANSACTION_FILE_PATH")
	if err != nil {
		return nil, err
	}

	azureBlobConnection := strings.TrimSpace(os.Getenv("AZURE_BLOB_CONNECTION_STRING"))
	azureBlobContainer := strings.TrimSpace(os.Getenv("AZURE_BLOB_CONTAINER"))
	reactClassificationModel := strings.TrimSpace(os.Getenv("REACT_CLASSIFICATION_MODEL"))
	if reactClassificationModel == "" {
		// Keep the default explicit here so prompt and retry tuning can change
		// independently from the surrounding transport settings.
		reactClassificationModel = "gpt-4o-mini"
	}
	reactClassificationSystemPrompt := defaultReactClassificationSystemPrompt
	reactClassificationBatchSize, err := optionalIntEnv("REACT_CLASSIFICATION_BATCH_SIZE", 10)
	if err != nil {
		return nil, err
	}
	reactClassificationFlushTimeout, err := optionalDurationEnv("REACT_CLASSIFICATION_FLUSH_TIMEOUT", 2*time.Second)
	if err != nil {
		return nil, err
	}
	reactClassificationClassifierFamily := strings.TrimSpace(os.Getenv("REACT_CLASSIFICATION_CLASSIFIER_FAMILY"))
	if reactClassificationClassifierFamily == "" {
		reactClassificationClassifierFamily = "react"
	}
	reactClassificationClassifierVersion := strings.TrimSpace(os.Getenv("REACT_CLASSIFICATION_CLASSIFIER_VERSION"))
	if reactClassificationClassifierVersion == "" {
		reactClassificationClassifierVersion = "v1"
	}
	reactClassificationPromptVersion := defaultReactClassificationPromptVersion
	reactClassificationRequestTimeout, err := optionalDurationEnv("REACT_CLASSIFICATION_REQUEST_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, err
	}
	reactClassificationMaxRetries, err := optionalIntEnv("REACT_CLASSIFICATION_MAX_RETRIES", 2)
	if err != nil {
		return nil, err
	}
	reactClassificationRetryBackoff, err := optionalDurationEnv("REACT_CLASSIFICATION_RETRY_BACKOFF", 2*time.Second)
	if err != nil {
		return nil, err
	}
	if reactClassificationBatchSize <= 0 {
		return nil, fmt.Errorf("REACT_CLASSIFICATION_BATCH_SIZE must be greater than 0")
	}
	if reactClassificationFlushTimeout <= 0 {
		return nil, fmt.Errorf("REACT_CLASSIFICATION_FLUSH_TIMEOUT must be greater than 0")
	}
	if reactClassificationMaxRetries < 0 {
		return nil, fmt.Errorf("REACT_CLASSIFICATION_MAX_RETRIES must be greater than or equal to 0")
	}
	if reactClassificationRetryBackoff < 0 {
		return nil, fmt.Errorf("REACT_CLASSIFICATION_RETRY_BACKOFF must be greater than or equal to 0")
	}

	return &AppConfig{
		Env:         env,
		ServiceName: serviceName,
		HTTP: HTTPConfig{
			Port:            port,
			ReadTimeout:     readTimeout,
			WriteTimeout:    writeTimeout,
			IdleTimeout:     idleTimeout,
			ShutdownTimeout: shutdownTimeout,
		},
		Database: DatabaseConfig{
			URL:            databaseURL,
			PingTimeout:    pingTimeout,
			MigrationsPath: filepath.Clean(migrationsPath),
		},
		Classification: ClassificationConfig{
			OpenAIAPIKey:           strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
			OpenAIBaseURL:          strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")),
			OpenAIAPIVersion:       strings.TrimSpace(os.Getenv("OPENAI_API_VERSION")),
			OpenAIUseAzure:         boolEnv("OPENAI_USE_AZURE"),
			ReactModel:             reactClassificationModel,
			ReactSystemPrompt:      reactClassificationSystemPrompt,
			ReactBatchSize:         reactClassificationBatchSize,
			ReactFlushTimeout:      reactClassificationFlushTimeout,
			ReactClassifierFamily:  reactClassificationClassifierFamily,
			ReactClassifierVersion: reactClassificationClassifierVersion,
			ReactPromptVersion:     reactClassificationPromptVersion,
			ReactRequestTimeout:    reactClassificationRequestTimeout,
			ReactMaxRetries:        reactClassificationMaxRetries,
			ReactRetryBackoff:      reactClassificationRetryBackoff,
		},
		Storage: StorageConfig{
			Provider:             strings.ToLower(strings.TrimSpace(storageProvider)),
			LocalTransactionPath: filepath.Clean(localTransactionPath),
			AzureBlobConnection:  azureBlobConnection,
			AzureBlobContainer:   azureBlobContainer,
		},
		Log: LogConfig{
			Level:       logLevel,
			FilePath:    filepath.Clean(logFilePath),
			ServiceName: serviceName,
			Environment: env,
		},
	}, nil
}

const defaultReactClassificationSystemPrompt = `You are an environmental data analyst classifying trade transactions against Paris Agreement alignment criteria.
For each transaction in the input batch, perform two independent checks, then make a final assessment. Confidence scores are integers 0–10: 0 = completely uncertain, 10 = fully confident.
---
## CHECK 1 — Not Aligned List
Determine whether the goods/service relates to ANY of these "not aligned" activities:
{% for entry in exclusionary_list %}
- {{ entry.activity_type }}
{% endfor %}
## CHECK 2 — Aligned List
Determine whether the goods/service fits ANY "aligned" activity in the table below, respecting stated conditions.
| Sector | Eligible operation type | Conditions and guidance |
| :--- | :--- | :--- |
{% for entry in u1_list %}
| {{ entry.sector }}  | {{ entry.eligible_operation_type }}  | {{ entry.condition_guidance }}  |
{% endfor %}
---
## FINAL ASSESSMENT
Given both checks, classify the transaction as ` + "`aligned`" + `, ` + "`not_aligned`" + `, or ` + "`next_step`" + `, and provide a one-sentence reason.
---
## OUTPUT SCHEMA
Return a single JSON array, one object per transaction, preserving input order. Emit no text outside the JSON block.
These field names are the internal classifier output contract and must be emitted exactly as written below.
The API later maps the two per-check confidence values into the frontend response as ` + "`step1_result.confidence`" + ` and ` + "`step2_result.confidence`" + `.
` + "```json" + `
[
  {
    "id": string,
    "not_aligned_list_match": bool,
    "not_aligned_list_match_confidence": int,
    "aligned_list_match": bool,
    "aligned_list_match_confidence": int,
    "overall_classification": "aligned" | "not_aligned" | "next_step",
    "reason": string
  }
]
` + "```" + `
---
## EXAMPLE
**Input:**
| ID | Description | Sector |
|---|---|---|
| GU2033-158-056 | CATO 306 MODIFIED CATIONIC STARCH | Food/Agr & Related Goods |
| GU2033-047-456 | SPANISH RAW COTTON | Food/Agr & Related Goods |
| GU2033-122-562 | INSTANT FILLED MILK POWDER | Food/Agr & Related Goods |
| GU2033-046-408 | FERROUS WASTE AND SCRAP | Raw/Non-Energy Com |
| GU2033-305-112 | THERMAL COAL | Energy & Related Goods |
**Output:**
` + "```json" + `
[
  {
    "id": "GU2033-158-056",
    "not_aligned_list_match": false,
    "not_aligned_list_match_confidence": 10,
    "aligned_list_match": false,
    "aligned_list_match_confidence": 7,
    "overall_classification": "next_step",
    "reason": "Modified cationic starch is a specialty chemical derivative not present on the not aligned list, but its end-use application cannot be confirmed as an aligned activity from the description alone."
  },
  {
    "id": "GU2033-047-456",
    "not_aligned_list_match": false,
    "not_aligned_list_match_confidence": 10,
    "aligned_list_match": false,
    "aligned_list_match_confidence": 6,
    "overall_classification": "next_step",
    "reason": "Raw cotton is an agricultural commodity that could qualify as low-GHG or climate-smart agriculture but requires verification of farming practices to confirm alignment."
  },
  {
    "id": "GU2033-122-562",
    "not_aligned_list_match": false,
    "not_aligned_list_match_confidence": 10,
    "aligned_list_match": false,
    "aligned_list_match_confidence": 9,
    "overall_classification": "next_step",
    "reason": "Filled milk powder is a ruminant livestock product and does not meet the non-ruminant low-GHG livestock condition required for alignment."
  },
  {
    "id": "GU2033-046-408",
    "not_aligned_list_match": false,
    "not_aligned_list_match_confidence": 10,
    "aligned_list_match": true,
    "aligned_list_match_confidence": 8,
    "overall_classification": "aligned",
    "reason": "Ferrous waste and scrap directly matches the aligned Waste sector activity for material recovery, supporting circular economy by reducing demand for primary metal production."
  },
  {
    "id": "GU2033-305-112",
    "not_aligned_list_match": true,
    "not_aligned_list_match_confidence": 10,
    "aligned_list_match": false,
    "aligned_list_match_confidence": 10,
    "overall_classification": "not_aligned",
    "reason": "Thermal coal is explicitly listed on the not aligned list."
  }
]`

const defaultReactClassificationPromptVersion = "v1"

func requiredEnv(key string) (string, error) {
	// Trim whitespace here so callers do not need to duplicate empty-string
	// handling across the large bootstrap sequence.
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return "", fmt.Errorf("%s is required", key)
	}

	return value, nil
}

func durationEnv(key string) (time.Duration, error) {
	value, err := requiredEnv(key)
	if err != nil {
		return 0, err
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parsing %s: %w", key, err)
	}

	return duration, nil
}

func optionalDurationEnv(key string, defaultValue time.Duration) (time.Duration, error) {
	// Optional durations still parse strictly when supplied so configuration typos
	// fail fast during startup.
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parsing %s: %w", key, err)
	}

	return duration, nil
}

func optionalIntEnv(key string, defaultValue int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}

	parsedValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parsing %s: %w", key, err)
	}

	return parsedValue, nil
}

func boolEnv(key string) bool {
	// Support the common truthy spellings used in shell environments and CI.
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	return value == "1" || value == "true" || value == "yes"
}
