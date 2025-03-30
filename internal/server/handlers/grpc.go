package handlers

import (
	"context"
	"fmt"
	"io"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/domain"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserIDKeyType is type for UserIDKey constant.
type UserIDKeyType string

const (
	UserIDKey UserIDKeyType = "user_id"

	NoUserIDLogMessage      = "No user ID"
	UserUnauthorizedMessage = "User unauthorized"
	InternalErrorMessage    = "Internal error"

	StatusUploaded  = "uploaded"
	StatusUploading = "uploading"
)

// UsecaseInterface contains the necessary functions for the business logic of app.
type UsecaseInterface interface {
	Register(ctx context.Context, login, password string, publicKey, privateKeyCipher []byte) (*domain.User, error)
	Login(ctx context.Context, login, password string) (*domain.User, error)

	CreateData(ctx context.Context, userID string, name string, dataType string, data []byte) (*domain.Data, error)
	CreateDataChunk(ctx context.Context, userID string, id string, chunk []byte) error
	SetDataStatus(ctx context.Context, userID string, id string, status string) error
	GetDataBatch(ctx context.Context, userID string, dataTypes []string) ([]*domain.Data, error)
	GetDataChunk(ctx context.Context, userID string, id string) ([]byte, error)
	UpdateData(ctx context.Context, userID string, id string, name string, dataType string, data []byte) (*domain.Data, error)
	DeleteData(ctx context.Context, userID string, id string) (*domain.Data, error)
}

// GRPCHandler handlers struct.
type GRPCHandler struct {
	pb.UnimplementedGoYandexGophkeeperServer

	usecase UsecaseInterface
}

// NewGRPCHandler creates *GRPCHandler
func NewGRPCHandler(usecase UsecaseInterface) *GRPCHandler {
	return &GRPCHandler{
		usecase: usecase,
	}
}

// getContextUserID gets user ID from context.
func getContextUserID(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("no context")
	}
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", fmt.Errorf("no %v", UserIDKey)
	}
	return userID, nil
}

// Register registers user.
func (h *GRPCHandler) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Register user")

	user, err := h.usecase.Register(ctx, in.Login, in.Password, in.PublicKey, in.PrivateKeyCipher)
	if err != nil {
		handlerLogger.Error("Failed to register user",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.RegisterResponse{
		PublicKey:        user.PublicKey,
		PrivateKeyCipher: user.PrivateKeyCipher,
	}, nil
}

// Login logins user.
func (h *GRPCHandler) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Login user")

	user, err := h.usecase.Login(ctx, in.Login, in.Password)
	if err != nil {
		handlerLogger.Error("Failed to register user",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.LoginResponse{
		PublicKey:        user.PublicKey,
		PrivateKeyCipher: user.PrivateKeyCipher,
	}, nil
}

// CreateData creates data.
func (h *GRPCHandler) CreateData(ctx context.Context, in *pb.CreateDataRequest) (*pb.CreateDataResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Create data")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	data, err := h.usecase.CreateData(
		ctx,
		userID,
		in.Name,
		in.Type,
		in.Data,
	)
	if err != nil {
		handlerLogger.Error("Failed to create data",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.CreateDataResponse{Id: data.ID}, nil
}

// GetDataBatch gets data batch.
func (h *GRPCHandler) GetDataBatch(ctx context.Context, in *pb.GetDataBatchRequest) (*pb.GetDataBatchResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Get data batch")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	batch, err := h.usecase.GetDataBatch(
		ctx,
		userID,
		in.Types,
	)
	if err != nil {
		handlerLogger.Error("Failed to get login/password batch",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	serializedBatch := make([]*pb.GetDataBatchResponse_Data, len(batch))
	for i, loginPassword := range batch {
		serializedBatch[i] = loginPassword.SerializeToProtobuf()
	}

	return &pb.GetDataBatchResponse{
		DataBatch: serializedBatch,
	}, nil
}

// UpdateData updates data.
func (h *GRPCHandler) UpdateData(ctx context.Context, in *pb.UpdateDataRequest) (*pb.UpdateDataResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Update data")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	_, err = h.usecase.UpdateData(
		ctx,
		userID,
		in.Id,
		in.Name,
		in.Type,
		in.Data,
	)
	if err != nil {
		handlerLogger.Error("Failed to update data",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.UpdateDataResponse{}, nil
}

// DeleteData deletes data.
func (h *GRPCHandler) DeleteData(ctx context.Context, in *pb.DeleteDataRequest) (*pb.DeleteDataResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Delete data")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	_, err = h.usecase.DeleteData(
		ctx,
		userID,
		in.Id,
	)
	if err != nil {
		handlerLogger.Error("Failed to delete data",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.DeleteDataResponse{}, nil
}

// CreateDataChunk creates data chunk.
func (h *GRPCHandler) CreateDataChunk(stream pb.GoYandexGophkeeper_CreateDataChunkServer) error {
	ctx := stream.Context()

	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Create data chunk")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	isFirstChunk := true

	for {
		chunk, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&pb.CreateDataChunkResponse{})
		}

		if err != nil {
			handlerLogger.Error("Failed to receive data chunk from stream",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}

		if isFirstChunk {
			err = h.usecase.SetDataStatus(ctx, userID, chunk.Id, StatusUploading)
			if err != nil {
				handlerLogger.Error("Failed to set data status",
					zap.Error(err),
				)
				return status.Error(codes.Internal, InternalErrorMessage)
			}

			isFirstChunk = false
		}

		if chunk.IsLast == true {
			err = h.usecase.SetDataStatus(ctx, userID, chunk.Id, StatusUploaded)
			if err != nil {
				handlerLogger.Error("Failed to set data status",
					zap.Error(err),
				)
				return status.Error(codes.Internal, InternalErrorMessage)
			}
		}

		err = h.usecase.CreateDataChunk(ctx, userID, chunk.Id, chunk.DataChunk)
		if err != nil {
			handlerLogger.Error("Failed to create data chunk",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}
	}
}

// GetDataChunk gets data chunk.
func (h *GRPCHandler) GetDataChunk(in *pb.GetDataChunkRequest, stream pb.GoYandexGophkeeper_GetDataChunkServer) error {
	ctx := stream.Context()

	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Get data chunk")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	for {
		chunk, err := h.usecase.GetDataChunk(ctx, userID, in.Id)

		if err == io.EOF {
			return nil
		}

		if err != nil {
			handlerLogger.Error("Failed to get data",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}

		err = stream.Send(&pb.GetDataChunkResponse{DataChunk: chunk})
		if err != nil {
			handlerLogger.Error("Failed to send data chunk in stream",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}
	}
}

// UpdateDataChunk updates data chunk.
func (h *GRPCHandler) UpdateDataChunk(stream pb.GoYandexGophkeeper_UpdateDataChunkServer) error {
	ctx := stream.Context()

	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Update data chunk")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	isFirstChunk := true

	for {
		chunk, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&pb.UpdateDataChunkResponse{})
		}

		if err != nil {
			handlerLogger.Error("Failed to receive data chunk from stream",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}

		if isFirstChunk {
			err = h.usecase.SetDataStatus(ctx, userID, chunk.Id, StatusUploading)
			if err != nil {
				handlerLogger.Error("Failed to set data status",
					zap.Error(err),
				)
				return status.Error(codes.Internal, InternalErrorMessage)
			}

			isFirstChunk = false
		}

		if chunk.IsLast == true {
			err = h.usecase.SetDataStatus(ctx, userID, chunk.Id, StatusUploaded)
			if err != nil {
				handlerLogger.Error("Failed to set data status",
					zap.Error(err),
				)
				return status.Error(codes.Internal, InternalErrorMessage)
			}
		}

		err = h.usecase.CreateDataChunk(ctx, userID, chunk.Id, chunk.DataChunk)
		if err != nil {
			handlerLogger.Error("Failed to create data chunk",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}
	}
}
