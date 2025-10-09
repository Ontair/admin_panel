package logger

import (
	"os"
	"path/filepath"

	"github.com/ontair/admin-panel/internal/core/ports/service"
	"github.com/ontair/admin-panel/internal/infra/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AppLogger implements port.Logger interface using zap
type AppLogger struct {
	logger *zap.Logger
}

// NewLogger creates new application logger
func NewLogger(cfg *config.Config) (service.Logger, error) {
	var (
		logger *zap.Logger
		err    error
	)

	if cfg.IsProduction() {
		// Production: JSON format
		config := zap.NewProductionConfig()
		if cfg.Logging.File != "" {
			config.OutputPaths = []string{cfg.Logging.File}
		}
		logger, err = config.Build()
	} else {
		// Development: Console format
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		logger, err = config.Build()
	}

	if err != nil {
		return nil, err
	}

	// Set log level if specified
	if cfg.Logging.Level != "" {
		var level zapcore.Level
		if err := level.Set(cfg.Logging.Level); err == nil {
			// This requires recreating the logger with the new level
			if cfg.IsProduction() {
				config := zap.NewProductionConfig()
				config.Level = zap.NewAtomicLevelAt(level)
				if cfg.Logging.File != "" {
					config.OutputPaths = []string{cfg.Logging.File}
				}
				logger, err = config.Build()
			} else {
				config := zap.NewDevelopmentConfig()
				config.Level = zap.NewAtomicLevelAt(level)
				logger, err = config.Build()
			}
			if err != nil {
				return nil, err
			}
		}
	}

	return &AppLogger{logger: logger}, nil
}

// Debug logs debug message
func (l *AppLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs info message
func (l *AppLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs warning message
func (l *AppLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs error message
func (l *AppLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs fatal message and exits
func (l *AppLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// Close closes the logger
func (l *AppLogger) Close() error {
	return l.logger.Sync()
}

// Ensure log directory exists
func EnsureLogDirectory(logFile string) error {
	if logFile != "" {
		dir := filepath.Dir(logFile)
		return os.MkdirAll(dir, 0755)
	}
	return nil
}
