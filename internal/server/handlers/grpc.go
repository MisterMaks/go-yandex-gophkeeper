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
	Register(ctx context.Context, login, password string) (*domain.User, error)
	Login(ctx context.Context, login, password string) (*domain.User, error)

	CreateLoginPassword(ctx context.Context, userID string, name, loginCipher, passwordCipher string) (*domain.LoginPassword, error)
	GetLoginPasswordBatch(ctx context.Context, userID string) ([]*domain.LoginPassword, error)
	UpdateLoginPassword(ctx context.Context, userID string, id string, name, loginCipher, passwordCipher string) (*domain.LoginPassword, error)
	DeleteLoginPassword(ctx context.Context, userID string, id string) (*domain.LoginPassword, error)

	CreateBankCard(ctx context.Context, userID string, name, numberCipher, expirationDateCipher, securityCodeCipher string) (*domain.BankCard, error)
	GetBankCardBatch(ctx context.Context, userID string) ([]*domain.BankCard, error)
	UpdateBankCard(ctx context.Context, userID string, id string, name, numberCipher, expirationDateCipher, securityCodeCipher string) (*domain.BankCard, error)
	DeleteBankCard(ctx context.Context, userID string, id string) (*domain.BankCard, error)

	CreateTextMeta(ctx context.Context, userID string, name string) (string, error)
	CreateText(ctx context.Context, userID string, id string, chunk string) error
	SetTextStatus(ctx context.Context, userID string, id string, status string) error
	GetTextMetaBatch(ctx context.Context, userID string) ([]*domain.TextMeta, error)
	GetText(ctx context.Context, userID string, id string) (string, error)
	UpdateTextMeta(ctx context.Context, userID string, id string, name string) error
	DeleteText(ctx context.Context, userID string, id string) (*domain.TextMeta, error)
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

	user, err := h.usecase.Register(ctx, in.Login, in.Password)
	if err != nil {
		handlerLogger.Error("Failed to register user",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.RegisterResponse{
		PublicKey:      user.PublicKey,
		PrivateKeyHash: user.PrivateKeyHash,
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
		PublicKey:      user.PublicKey,
		PrivateKeyHash: user.PrivateKeyHash,
	}, nil
}

// CreateLoginPassword creates login/password.
func (h *GRPCHandler) CreateLoginPassword(ctx context.Context, in *pb.CreateLoginPasswordRequest) (*pb.CreateLoginPasswordResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Create login/password")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	_, err = h.usecase.CreateLoginPassword(
		ctx,
		userID,
		in.Name,
		in.LoginCipher,
		in.PasswordCipher,
	)
	if err != nil {
		handlerLogger.Error("Failed to create login/password",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.CreateLoginPasswordResponse{}, nil
}

// GetLoginPasswordBatch gets login/password batch.
func (h *GRPCHandler) GetLoginPasswordBatch(ctx context.Context, _ *pb.GetLoginPasswordBatchRequest) (*pb.GetLoginPasswordBatchResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Get login/password batch")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	batch, err := h.usecase.GetLoginPasswordBatch(
		ctx,
		userID,
	)
	if err != nil {
		handlerLogger.Error("Failed to get login/password batch",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	serializedBatch := make([]*pb.GetLoginPasswordBatchResponse_LoginPassword, len(batch))
	for i, loginPassword := range batch {
		serializedBatch[i] = loginPassword.SerializeToProtobuf()
	}

	return &pb.GetLoginPasswordBatchResponse{
		LoginPasswordBatch: serializedBatch,
	}, nil
}

// UpdateLoginPassword updates login/password.
func (h *GRPCHandler) UpdateLoginPassword(ctx context.Context, in *pb.UpdateLoginPasswordRequest) (*pb.UpdateLoginPasswordResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Update login/password")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	_, err = h.usecase.UpdateLoginPassword(
		ctx,
		userID,
		in.Id,
		in.Name,
		in.LoginCipher,
		in.PasswordCipher,
	)
	if err != nil {
		handlerLogger.Error("Failed to update login/password",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.UpdateLoginPasswordResponse{}, nil
}

// DeleteLoginPassword deletes login/password.
func (h *GRPCHandler) DeleteLoginPassword(ctx context.Context, in *pb.DeleteLoginPasswordRequest) (*pb.DeleteLoginPasswordResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Delete login/password")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	_, err = h.usecase.DeleteLoginPassword(
		ctx,
		userID,
		in.Id,
	)
	if err != nil {
		handlerLogger.Error("Failed to delete login/password",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.DeleteLoginPasswordResponse{}, nil
}

// CreateBankCard creates bank card.
func (h *GRPCHandler) CreateBankCard(ctx context.Context, in *pb.CreateBankCardRequest) (*pb.CreateBankCardResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Create bank card")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	_, err = h.usecase.CreateBankCard(
		ctx,
		userID,
		in.Name,
		in.NumberCipher,
		in.ExpirationDateCipher,
		in.SecurityCodeCipher,
	)
	if err != nil {
		handlerLogger.Error("Failed to create bank card",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.CreateBankCardResponse{}, nil
}

// GetBankCardBatch gets bank card batch.
func (h *GRPCHandler) GetBankCardBatch(ctx context.Context, _ *pb.GetBankCardBatchRequest) (*pb.GetBankCardBatchResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Get bank card batch")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	batch, err := h.usecase.GetBankCardBatch(
		ctx,
		userID,
	)
	if err != nil {
		handlerLogger.Error("Failed to get bank card batch",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	serializedBatch := make([]*pb.GetBankCardBatchResponse_BankCard, len(batch))
	for i, bankCard := range batch {
		serializedBatch[i] = bankCard.SerializeToProtobuf()
	}

	return &pb.GetBankCardBatchResponse{
		BankCardBatch: serializedBatch,
	}, nil
}

// UpdateBankCard updates bank card.
func (h *GRPCHandler) UpdateBankCard(ctx context.Context, in *pb.UpdateBankCardRequest) (*pb.UpdateBankCardResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Update bank card")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	_, err = h.usecase.UpdateBankCard(
		ctx,
		userID,
		in.Id,
		in.Name,
		in.NumberCipher,
		in.ExpirationDateCipher,
		in.SecurityCodeCipher,
	)
	if err != nil {
		handlerLogger.Error("Failed to update bank card",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.UpdateBankCardResponse{}, nil
}

// DeleteBankCard deletes bank card.
func (h *GRPCHandler) DeleteBankCard(ctx context.Context, in *pb.DeleteBankCardRequest) (*pb.DeleteBankCardResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Delete bank card")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	_, err = h.usecase.DeleteBankCard(
		ctx,
		userID,
		in.Id,
	)
	if err != nil {
		handlerLogger.Error("Failed to delete bank card",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.DeleteBankCardResponse{}, nil
}

// CreateTextMeta creates text meta.
func (h *GRPCHandler) CreateTextMeta(ctx context.Context, in *pb.CreateTextMetaRequest) (*pb.CreateTextMetaResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Create text meta")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	id, err := h.usecase.CreateTextMeta(
		ctx,
		userID,
		in.Name,
	)
	if err != nil {
		handlerLogger.Error("Failed to create text meta",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.CreateTextMetaResponse{Id: id}, nil
}

// CreateText creates text.
func (h *GRPCHandler) CreateText(stream pb.GoYandexGophkeeper_CreateTextServer) error {
	ctx := stream.Context()

	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Create text")

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
			return stream.SendAndClose(&pb.CreateTextResponse{})
		}

		if err != nil {
			handlerLogger.Error("Failed to receive text from stream",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}

		if isFirstChunk {
			err = h.usecase.SetTextStatus(ctx, userID, chunk.Id, StatusUploading)
			if err != nil {
				handlerLogger.Error("Failed to set text uploaded",
					zap.Error(err),
				)
				return status.Error(codes.Internal, InternalErrorMessage)
			}

			isFirstChunk = false
		}

		if chunk.IsFinished != nil && *chunk.IsFinished == true {
			err = h.usecase.SetTextStatus(ctx, userID, chunk.Id, StatusUploaded)
			if err != nil {
				handlerLogger.Error("Failed to set text uploaded",
					zap.Error(err),
				)
				return status.Error(codes.Internal, InternalErrorMessage)
			}
		}

		err = h.usecase.CreateText(ctx, userID, chunk.Id, chunk.TextChunk)
		if err != nil {
			handlerLogger.Error("Failed to create text",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}
	}
}

// GetTextMetaBatch gets text meta batch.
func (h *GRPCHandler) GetTextMetaBatch(ctx context.Context, _ *pb.GetTextMetaBatchRequest) (*pb.GetTextMetaBatchResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Get text meta batch")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	batch, err := h.usecase.GetTextMetaBatch(
		ctx,
		userID,
	)
	if err != nil {
		handlerLogger.Error("Failed to get text meta batch",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	serializedBatch := make([]*pb.GetTextMetaBatchResponse_TextMeta, len(batch))
	for i, textMeta := range batch {
		serializedBatch[i] = textMeta.SerializeToProtobuf()
	}

	return &pb.GetTextMetaBatchResponse{
		TextMetaBatch: serializedBatch,
	}, nil
}

// GetText gets text.
func (h *GRPCHandler) GetText(in *pb.GetTextRequest, stream pb.GoYandexGophkeeper_GetTextServer) error {
	ctx := stream.Context()

	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Get text")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	for {
		chunk, err := h.usecase.GetText(ctx, userID, in.Id)

		if err == io.EOF {
			return nil
		}

		if err != nil {
			handlerLogger.Error("Failed to get text",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}

		err = stream.Send(&pb.GetTextResponse{TextChunk: chunk})
		if err != nil {
			handlerLogger.Error("Failed to send text in stream",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}
	}
}

// UpdateTextMeta updates text meta.
func (h *GRPCHandler) UpdateTextMeta(ctx context.Context, in *pb.UpdateTextMetaRequest) (*pb.UpdateTextMetaResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Update text meta")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	err = h.usecase.UpdateTextMeta(
		ctx,
		userID,
		in.Id,
		in.Name,
	)
	if err != nil {
		handlerLogger.Error("Failed to update text meta",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.UpdateTextMetaResponse{}, nil
}

// UpdateText updates text.
func (h *GRPCHandler) UpdateText(stream pb.GoYandexGophkeeper_UpdateTextServer) error {
	ctx := stream.Context()

	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("update text")

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
			return stream.SendAndClose(&pb.UpdateTextResponse{})
		}

		if err != nil {
			handlerLogger.Error("Failed to receive text from stream",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}

		if isFirstChunk {
			err = h.usecase.SetTextStatus(ctx, userID, chunk.Id, StatusUploading)
			if err != nil {
				handlerLogger.Error("Failed to set text uploaded",
					zap.Error(err),
				)
				return status.Error(codes.Internal, InternalErrorMessage)
			}

			isFirstChunk = false
		}

		if chunk.IsFinished != nil && *chunk.IsFinished == true {
			err = h.usecase.SetTextStatus(ctx, userID, chunk.Id, StatusUploaded)
			if err != nil {
				handlerLogger.Error("Failed to set text uploaded",
					zap.Error(err),
				)
				return status.Error(codes.Internal, InternalErrorMessage)
			}
		}

		err = h.usecase.CreateText(ctx, userID, chunk.Id, chunk.TextChunk)
		if err != nil {
			handlerLogger.Error("Failed to create text",
				zap.Error(err),
			)
			return status.Error(codes.Internal, InternalErrorMessage)
		}
	}
}

// DeleteText deletes text.
func (h *GRPCHandler) DeleteText(ctx context.Context, in *pb.DeleteTextRequest) (*pb.DeleteTextResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Delete text")

	userID, err := getContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn(NoUserIDLogMessage,
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, UserUnauthorizedMessage)
	}

	_, err = h.usecase.DeleteText(
		ctx,
		userID,
		in.Id,
	)
	if err != nil {
		handlerLogger.Error("Failed to delete text",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, InternalErrorMessage)
	}

	return &pb.DeleteTextResponse{}, nil
}
