package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/domain"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/logger"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UserIDKeyType is type for UserIDKey constant.
type UserIDKeyType string

const (
	UserIDKey              UserIDKeyType = "user_id"
	AccessTokenKey         string        = "accessToken"
	AuthorizationHeaderKey               = "Authorization"
	BearerKey                            = "Bearer "

	NoUserIDLogMessage      = "No user ID"
	UserUnauthorizedMessage = "User unauthorized"
	InternalErrorMessage    = "Internal error"

	ChunkSize uint = 100 * 1024
)

// UsecaseInterface contains the necessary functions for the business logic of app.
type UsecaseInterface interface {
	Register(ctx context.Context, login, password string, publicKey, privateKeyCipher []byte) (*domain.User, error)
	Login(ctx context.Context, login, password string) (*domain.User, error)

	CreateData(ctx context.Context, userID string, name string, dataType string, data []byte) (*domain.Data, error)
	GetDataBatch(ctx context.Context, userID string) ([]*domain.Data, error)
	UpdateData(ctx context.Context, userID string, id string, name string, dataType string, data []byte) (*domain.Data, error)
	DeleteData(ctx context.Context, userID string, id string) (*domain.Data, error)

	//CreateChunkedData(userID string, stream pb.GoYandexGophkeeper_CreateChunkedDataServer) (*domain.Data, error)
	//GetChunkedDataReader(ctx context.Context, userID string, id string) (io.Reader, error)
	//UpdateChunkedData(userID string, stream pb.GoYandexGophkeeper_UpdateChunkedDataServer) (*domain.Data, error)
}

// GRPCHandler handlers struct.
type GRPCHandler struct {
	pb.UnimplementedGoYandexGophkeeperServer

	usecase UsecaseInterface

	tokenKey string
	tokenExp time.Duration

	grpcMethodsForAuthenticateUnaryInterceptor map[string]struct{}
}

// NewGRPCHandler creates *GRPCHandler
func NewGRPCHandler(
	usecase UsecaseInterface,
	tokenKey string,
	tokenExp time.Duration,
	grpcMethodsForAuthenticateUnaryInterceptorSl []string,
) *GRPCHandler {
	grpcMethodsForAuthenticateUnaryInterceptor := map[string]struct{}{}
	for _, grpcMethod := range grpcMethodsForAuthenticateUnaryInterceptorSl {
		grpcMethodsForAuthenticateUnaryInterceptor[grpcMethod] = struct{}{}
	}

	return &GRPCHandler{
		usecase:  usecase,
		tokenKey: tokenKey,
		tokenExp: tokenExp,
		grpcMethodsForAuthenticateUnaryInterceptor: grpcMethodsForAuthenticateUnaryInterceptor,
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

// buildJWTString creates token and return it in string format.
func (h *GRPCHandler) buildJWTString(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.tokenExp)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(h.tokenKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Register registers user.
func (h *GRPCHandler) Register(ctx context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Register user")

	user, err := h.usecase.Register(ctx, in.GetLogin(), in.GetLogin(), in.GetPublicKey(), in.GetPrivateKeyCipher())

	switch err {
	case nil:
	case domain.ErrLoginTaken:
		return nil, status.Error(codes.AlreadyExists, err.Error())
	case domain.ErrInvalidLoginPasswordFormat:
		return nil, status.Error(codes.InvalidArgument, err.Error())
	default:
		handlerLogger.Error("Failed to register user", zap.Error(err))
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	accessToken, err := h.buildJWTString(user.ID)
	if err != nil {
		handlerLogger.Error("Failed to build JWT string", zap.Error(err))
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.RegisterResponse{
		AccessToken: accessToken,
	}, nil
}

// Login logins user.
func (h *GRPCHandler) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Login user")

	user, err := h.usecase.Login(ctx, in.GetLogin(), in.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrInvalidLoginPassword) {
			handlerLogger.Warn("Failed to login user",
				zap.Error(err),
			)
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		handlerLogger.Error("Failed to login user",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	accessToken, err := h.buildJWTString(user.ID)
	if err != nil {
		handlerLogger.Error("Failed to build JWT string", zap.Error(err))
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.LoginResponse{
		PublicKey:        user.PublicKey,
		PrivateKeyCipher: user.PrivateKeyCipher,
		AccessToken:      accessToken,
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
		in.GetName(),
		in.GetType().String(),
		in.GetData(),
	)
	if err != nil {
		if errors.Is(err, domain.ErrDataWithThisNameAndTypeExists) {
			handlerLogger.Warn("Failed to create data",
				zap.Error(err),
			)
			return nil, status.Error(codes.AlreadyExists, "data with this name and type already exists")
		}
		handlerLogger.Error("Failed to create data",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.CreateDataResponse{Id: data.ID}, nil
}

// GetDataBatch gets data batch.
func (h *GRPCHandler) GetDataBatch(ctx context.Context, _ *pb.GetDataBatchRequest) (*pb.GetDataBatchResponse, error) {
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
		in.GetId(),
		in.GetName(),
		in.GetType().String(),
		in.GetData(),
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
		in.GetId(),
	)
	if err != nil {
		handlerLogger.Error("Failed to delete data",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.DeleteDataResponse{}, nil
}

// CreateChunkedData creates chunked data.
//func (h *GRPCHandler) CreateChunkedData(stream pb.GoYandexGophkeeper_CreateChunkedDataServer) error {
//	ctx := stream.Context()
//
//	handlerLogger := logger.GetContextLogger(ctx)
//
//	handlerLogger.Info("Create data chunk")
//
//	userID, err := getContextUserID(ctx)
//	if err != nil {
//		handlerLogger.Warn(NoUserIDLogMessage,
//			zap.Error(err),
//		)
//		return status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
//	}
//
//	metaData, err := h.usecase.CreateChunkedData(userID, stream)
//	if err != nil {
//		handlerLogger.Error("Failed to create chunked data",
//			zap.Error(err),
//		)
//
//		return status.Error(codes.Internal, InternalErrorMessage)
//	}
//
//	return stream.SendAndClose(&pb.CreateChunkedDataResponse{Id: metaData.ID})
//}

// GetChunkedData gets data chunk.
//func (h *GRPCHandler) GetChunkedData(in *pb.GetDataChunkRequest, stream pb.GoYandexGophkeeper_GetChunkedDataServer) error {
//	ctx := stream.Context()
//
//	handlerLogger := logger.GetContextLogger(ctx)
//
//	handlerLogger.Info("Get data chunk")
//
//	userID, err := getContextUserID(ctx)
//	if err != nil {
//		handlerLogger.Warn(NoUserIDLogMessage,
//			zap.Error(err),
//		)
//		return status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
//	}
//
//	reader, err := h.usecase.GetChunkedDataReader(ctx, userID, in.Id)
//
//	for {
//		buffer := make([]byte, ChunkSize)
//
//		n, err := reader.Read(buffer)
//		if err == io.EOF {
//			return nil
//		}
//
//		if err != nil {
//			handlerLogger.Error("Failed to get data",
//				zap.Error(err),
//			)
//			return status.Error(codes.Internal, InternalErrorMessage)
//		}
//
//		err = stream.Send(&pb.GetDataChunkResponse{DataChunk: buffer[:n]})
//		if err != nil {
//			handlerLogger.Error("Failed to send data chunk in stream",
//				zap.Error(err),
//			)
//			return status.Error(codes.Internal, InternalErrorMessage)
//		}
//	}
//}

// UpdateChunkedData updates data chunk.
//func (h *GRPCHandler) UpdateChunkedData(stream pb.GoYandexGophkeeper_UpdateChunkedDataServer) error {
//	ctx := stream.Context()
//
//	handlerLogger := logger.GetContextLogger(ctx)
//
//	handlerLogger.Info("Update data chunk")
//
//	userID, err := getContextUserID(ctx)
//	if err != nil {
//		handlerLogger.Warn(NoUserIDLogMessage,
//			zap.Error(err),
//		)
//		return status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
//	}
//
//	_, err = h.usecase.UpdateChunkedData(userID, stream)
//	if err != nil {
//		handlerLogger.Error("Failed to update chunked data",
//			zap.Error(err),
//		)
//
//		return status.Error(codes.Internal, InternalErrorMessage)
//	}
//
//	return stream.SendAndClose(&pb.UpdateChunkedDataResponse{})
//}
