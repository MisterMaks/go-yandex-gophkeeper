package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		level   string
		wantErr bool
	}{
		{
			name:    "info",
			level:   "info",
			wantErr: false,
		},
		{
			name:    "warn",
			level:   "warn",
			wantErr: false,
		},
		{
			name:    "debug",
			level:   "debug",
			wantErr: false,
		},
		{
			name:    "error",
			level:   "error",
			wantErr: false,
		},
		{
			name:    "invalid",
			level:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.level)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestGetContextLogger(t *testing.T) {
	l := Log.With(zap.String("key", "value"))

	tests := []struct {
		name   string
		ctx    context.Context
		logger *zap.Logger
	}{
		{
			name:   "default logger",
			ctx:    context.Background(),
			logger: Log,
		},
		{
			name:   "context logger",
			ctx:    context.WithValue(context.Background(), LoggerKey, l),
			logger: l,
		},
		{
			name:   "invalid logger",
			ctx:    context.WithValue(context.Background(), LoggerKey, "invalid logger"),
			logger: Log,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualLogger := GetContextLogger(tt.ctx)
			assert.Equal(t, tt.logger, actualLogger)
		})
	}
}
