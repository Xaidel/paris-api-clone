package observability

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gyud-adb/paris-api/internal/infrastructure/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger builds the application logger.
func NewLogger(cfg config.LogConfig) (*zap.Logger, func() error, error) {
	level, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing log level: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0o755); err != nil {
		return nil, nil, fmt.Errorf("creating log directory: %w", err)
	}

	file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("opening log file: %w", err)
	}

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(consoleEncoderConfig()),
		zapcore.Lock(os.Stdout),
		level,
	)
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(fileEncoderConfig()),
		zapcore.AddSync(file),
		level,
	)

	logger := zap.New(zapcore.NewTee(consoleCore, fileCore), zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel)).With(
		zap.String("service", cfg.ServiceName),
		zap.String("environment", cfg.Environment),
	)

	cleanup := func() error {
		_ = logger.Sync()
		return file.Close()
	}

	return logger, cleanup, nil
}

func consoleEncoderConfig() zapcore.EncoderConfig {
	config := zap.NewDevelopmentEncoderConfig()
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.TimeKey = "timestamp"

	return config
}

func fileEncoderConfig() zapcore.EncoderConfig {
	config := zap.NewProductionEncoderConfig()
	config.TimeKey = "timestamp"
	config.MessageKey = "message"
	config.LevelKey = "severity"
	config.NameKey = "logger"
	config.CallerKey = "caller"
	config.StacktraceKey = "stacktrace"
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.LowercaseLevelEncoder

	return config
}
