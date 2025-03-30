package server

import (
	"context"
	"database/sql"
	"log"

	internal_config "github.com/MisterMaks/go-yandex-gophkeeper/internal/server/config"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/logger"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

const (
	ConfigKey      = "config"
	GRPCAddressKey = "grpc_address"
)

func migrate(dsn string) error {
	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		return err
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Log.Fatal("Failed to close DB",
				zap.Error(err),
			)
		}
	}()

	ctx := context.Background()

	return goose.RunContext(ctx, "up", db, "./migrations/postgres/")
}

func connectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		logger.Log.Error("Failed to ping DB Postgres",
			zap.Error(err),
		)
	}
	return db, nil
}

func main() {
	config, err := internal_config.New()
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to create config. Error:", err)
	}

	err = logger.New(config.LogLevel)
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to init logger. Error:", err)
	}

	logger.Log.Debug("Config data",
		zap.Any(ConfigKey, config),
	)

	logger.Log.Info("Applying migrations")
	err = migrate(config.PostgresDSN)
	if err != nil {
		logger.Log.Fatal("Failed to apply migrations",
			zap.Error(err),
		)
	}

	postgresDB, err := connectPostgres(config.PostgresDSN)
	if err != nil {
		logger.Log.Fatal("Failed to connect to Postgres",
			zap.Error(err),
		)
	}
	defer func() {
		err = postgresDB.Close()
		if err != nil {
			logger.Log.Error("Failed to close Postgres",
				zap.Error(err),
			)
		}
	}()
}
