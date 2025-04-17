package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is global logger.
var Log *zap.Logger = zap.NewNop()

// LoggerKeyType is type for LoggerKey constant.
type LoggerKeyType string

// Constants for logger.
const ()

// New init logger.
func New(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}
