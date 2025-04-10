package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	internal_config "github.com/MisterMaks/go-yandex-gophkeeper/internal/server/config"
	internal_handler "github.com/MisterMaks/go-yandex-gophkeeper/internal/server/handler"
	internal_db "github.com/MisterMaks/go-yandex-gophkeeper/internal/server/infrastructure/db"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/logger"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/usecase"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	ConfigKey         = "config"
	AddressKey        = "address"
	PathToCertificate = "certificates/cert.pem"
	PathToPrivateKey  = "certificates/private_key.pem"
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

	postgresStorage := internal_db.NewPostgresStorage(postgresDB)

	//minioClient, err := minio.New(config.MinioEndpoint, &minio.Options{
	//	Creds: credentials.NewStaticV4(config.MinioAccessKey, config.MinioSecretKey, ""),
	//})
	//if err != nil {
	//	logger.Log.Fatal("Failed to create Minio client",
	//		zap.Error(err),
	//	)
	//}

	//minioStorage := internal_db.NewMinioStorage(minioClient)

	u := usecase.NewUsecase(
		postgresStorage,
		nil,
		config.PasswordKey,
		config.MinLoginLength,
		config.MinPasswordLength,
	)

	handler := internal_handler.NewGRPCHandler(
		u,
		config.TokenKey,
		config.TokenExpiration,
		[]string{
			pb.GoYandexGophkeeper_CreateData_FullMethodName,
			pb.GoYandexGophkeeper_GetDataBatch_FullMethodName,
			pb.GoYandexGophkeeper_UpdateData_FullMethodName,
			pb.GoYandexGophkeeper_DeleteData_FullMethodName,
		},
	)

	tlsCert, err := credentials.NewServerTLSFromFile(PathToCertificate, PathToPrivateKey)
	if err != nil {
		logger.Log.Fatal("Failed to get server certificate from file",
			zap.Error(err),
		)
	}

	server := grpc.NewServer(
		grpc.Creds(tlsCert),
		grpc.ChainUnaryInterceptor(
			logger.RequestLoggerUnaryInterceptor,
			handler.AuthenticateUnaryInterceptor,
		),
	)

	pb.RegisterGoYandexGophkeeperServer(server, handler)

	listen, err := net.Listen("tcp", config.GRPCAddress)
	if err != nil {
		logger.Log.Fatal("Failed to create listen",
			zap.Error(err),
		)
	}

	logger.Log.Info("Server running",
		zap.String(AddressKey, config.GRPCAddress),
	)

	go func() {
		err = server.Serve(listen)
		if err != nil && err != grpc.ErrServerStopped {
			logger.Log.Fatal("Failed to start GRPC server",
				zap.Error(err),
			)
		}
	}()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	exitSyg := <-exitChan
	logger.Log.Info("terminating: via signal", zap.Any("signal", exitSyg))
	server.GracefulStop()
}
