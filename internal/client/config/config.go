package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/spf13/viper"
)

const AppName = "go-yandex-gophkeeper-client"

// Config config data for app.
type Config struct {
	AppName        string
	GRPCAddress    string `env:"GRPC_ADDRESS" mapstructure:"grpc_address"`
	ConfigFilePath string `env:"CONFIG_FILE_PATH"`
}

var DefaultConfig = map[string]interface{}{
	"grpc_address": "127.0.0.1:8080",
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
