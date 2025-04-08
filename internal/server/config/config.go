package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/spf13/viper"
)

const AppName = "go-yandex-gophkeeper-server"

// Config config data for app.
type Config struct {
	AppName     string
	GRPCAddress string `env:"GRPC_ADDRESS" mapstructure:"grpc_address"`

	PasswordKey       string `env:"PASSWORD_KEY" mapstructure:"password_key"`
	MinLoginLength    int    `env:"MIN_LOGIN_LENGTH" mapstructure:"min_login_length"`
	MinPasswordLength int    `env:"MIN_PASSWORD_LENGTH" mapstructure:"min_password_length"`

	TokenKey        string        `env:"TOKEN_KEY" mapstructure:"token_key"`
	TokenExpiration time.Duration `env:"TOKEN_EXPIRATION" mapstructure:"token_expiration"`

	LogLevel string `env:"LOG_LEVEL" mapstructure:"log_level"`

	PostgresDSN string `env:"POSTGRES_DSN" mapstructure:"postgres_dsn"`

	MinioEndpoint   string `env:"MINIO_ENDPOINT" mapstructure:"minio_endpoint"`
	MinioAccessKey  string `env:"MINIO_ACCESS_KEY" mapstructure:"minio_access_key"`
	MinioSecretKey  string `env:"MINIO_SECRET_KEY" mapstructure:"minio_secret_key"`
	MinioBucketName string `env:"MINIO_BUCKET_NAME" mapstructure:"minio_bucket_name"`

	ConfigFilePath string `env:"CONFIG_FILE_PATH"`
}

var DefaultConfig = map[string]interface{}{
	"grpc_address":        "localhost:8081",
	"password_key":        "",
	"min_login_length":    0,
	"min_password_length": 0,
	"token_key":           "",
	"token_expiration":    time.Minute,
	"log_level":           "INFO",
	"postgres_dsn":        "",
	"minio_endpoint":      "",
	"minio_access_key":    "",
	"minio_secret_key":    "",
	"minio_bucket_name":   "",
}

// New create config.
func New() (*Config, error) {
	c := &Config{
		AppName: AppName,
	}

	err := env.Parse(c)
	if err != nil {
		return nil, err
	}

	v := viper.New()

	for key, value := range DefaultConfig {
		v.SetDefault(key, value)
	}

	v.SetConfigName(AppName)
	v.AutomaticEnv()

	if c.ConfigFilePath != "" {
		v.SetConfigFile(c.ConfigFilePath)

		err = v.ReadInConfig()
		if err != nil {
			return nil, err
		}
	}

	err = v.Unmarshal(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
