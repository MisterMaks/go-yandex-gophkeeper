package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	expectedConfig := &Config{
		AppName:         "go-yandex-gophkeeper-server",
		GRPCAddress:     "127.0.0.1:8080",
		TokenExpiration: time.Minute,
		LogLevel:        "INFO",
	}

	config, err := New()

	assert.NoError(t, err)
	assert.Equal(t, expectedConfig, config)
}
