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
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/config"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/handler"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/infrastructure/db"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/logger"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/usecase"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/minio/minio-go/v7"
	minio_credentials "github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpc_credentials "google.golang.org/grpc/credentials"
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
	c, err := config.New()
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to create config. Error:", err)
	}

	err = logger.New(c.LogLevel)
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to init logger. Error:", err)
	}

	logger.Log.Debug("Config data",
		zap.Any(ConfigKey, c),
	)

	logger.Log.Info("Applying migrations")
	err = migrate(c.PostgresDSN)
	if err != nil {
		logger.Log.Fatal("Failed to apply migrations",
			zap.Error(err),
		)
	}

	postgresDB, err := connectPostgres(c.PostgresDSN)
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

	postgresStorage := db.NewPostgresStorage(postgresDB)

	minioClient, err := minio.New(c.MinioEndpoint, &minio.Options{
		Creds: minio_credentials.NewStaticV4(c.MinioAccessKey, c.MinioSecretKey, ""),
	})
	if err != nil {
		logger.Log.Fatal("Failed to create Minio client",
			zap.Error(err),
		)
	}

	minioStorage := db.NewMinioStorage(minioClient)

	u := usecase.NewUsecase(
		postgresStorage,
		minioStorage,
		c.PasswordKey,
		c.MinLoginLength,
		c.MinPasswordLength,
	)

	h := handler.NewGRPCHandler(
		u,
		c.TokenKey,
		c.TokenExpiration,
		[]string{
			pb.GoYandexGophkeeper_CreateData_FullMethodName,
			pb.GoYandexGophkeeper_GetDataBatch_FullMethodName,
			pb.GoYandexGophkeeper_UpdateData_FullMethodName,
			pb.GoYandexGophkeeper_DeleteData_FullMethodName,
		},
		[]string{
			pb.GoYandexGophkeeper_CreateChunkedData_FullMethodName,
			pb.GoYandexGophkeeper_GetChunkedData_FullMethodName,
		},
	)

	tlsCert, err := grpc_credentials.NewServerTLSFromFile(PathToCertificate, PathToPrivateKey)
	if err != nil {
		logger.Log.Fatal("Failed to get server certificate from file",
			zap.Error(err),
		)
	}

	server := grpc.NewServer(
		grpc.Creds(tlsCert),
		grpc.ChainUnaryInterceptor(
			logger.RequestLoggerUnaryInterceptor,
			h.AuthenticateUnaryInterceptor,
		),
		grpc.StreamInterceptor(
			h.AuthenticateStreamInterceptor,
		),
	)

	pb.RegisterGoYandexGophkeeperServer(server, h)

	listen, err := net.Listen("tcp", c.GRPCAddress)
	if err != nil {
		logger.Log.Fatal("Failed to create listen",
			zap.Error(err),
		)
	}

	logger.Log.Info("Server running",
		zap.String(AddressKey, c.GRPCAddress),
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
