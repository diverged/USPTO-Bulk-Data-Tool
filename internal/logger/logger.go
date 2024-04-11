package logger

import (
	"fmt"
	"time"

	// "github.com/diverged/uspt-go/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/diverged/uspto-bulk-data-tool/internal/config"

	parser "github.com/diverged/uspt-go/types"
)

func InitLogger(cfg config.LoggerConfig /* cfg config.LoggerConfig */) (*zap.Logger, error) {
	var level zapcore.Level
	switch cfg.LogLevel {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.ErrorLevel
	}

	// Name the log file with a timestamp unique to the particular execution

	// * timestamped logging
	timestamp := time.Now().Format("15:04:05")
	logFileName := fmt.Sprintf("log-%s.jsonl", timestamp)

	developmentMode := cfg.LogMode == "dev"

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: developmentMode,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.LowercaseLevelEncoder,
			// EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeTime:     CustomTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout", fmt.Sprintf("%s/%s", cfg.LogPath, logFileName)},
		ErrorOutputPaths: []string{"stderr"},
	}

	if developmentMode {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Colorize logs in development mode
		config.OutputPaths = []string{"stdout"}                             // Log only to stdout in development mode
		config.ErrorOutputPaths = []string{"stderr"}
	}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// * Implement the uspt-go interface
	var _ parser.Logger = (*ZapLoggerAdapter)(nil)

	zap.ReplaceGlobals(logger)
	return logger, nil

}

func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("15:04:05.000"))
}

// * Interface adapter for connecting zap with uspt-go logger

// ZapLoggerAdapter adapts a zap.Logger to the parser.Logger interface.
type ZapLoggerAdapter struct {
	logger *zap.Logger
}

// NewZapLoggerAdapter creates a new ZapLoggerAdapter.
func NewZapLoggerAdapter(logger *zap.Logger) *ZapLoggerAdapter {
	return &ZapLoggerAdapter{logger: logger}
}

// Debug logs a debug message with additional key-value pairs.
func (z *ZapLoggerAdapter) Debug(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Debugw(msg, keysAndValues...)
}

// Info logs an info message with additional key-value pairs.
func (z *ZapLoggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Infow(msg, keysAndValues...)
}

// Warn logs a warning message with additional key-value pairs.
func (z *ZapLoggerAdapter) Warn(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Warnw(msg, keysAndValues...)
}

// Error logs an error message with additional key-value pairs.
func (z *ZapLoggerAdapter) Error(msg string, keysAndValues ...interface{}) {
	z.logger.Sugar().Errorw(msg, keysAndValues...)
}
