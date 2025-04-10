package handler

import (
	"context"
	"errors"
	"testing"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/domain"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/handler/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewGRPCHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUsecaseInterface(ctrl)

	tokenKey := "token_key"
	tokenExp := time.Second

	methodName := "MethodName"

	grpcHandler := &GRPCHandler{
		usecase:  mockUsecase,
		tokenKey: tokenKey,
		tokenExp: tokenExp,
		grpcMethodsForAuthenticateUnaryInterceptor: map[string]struct{}{methodName: {}},
	}

	actualGRPCHandler := NewGRPCHandler(mockUsecase, tokenKey, tokenExp, []string{methodName})

	assert.Equal(t, grpcHandler, actualGRPCHandler)
}

func TestGRPCHandler_Register(t *testing.T) {
	login := "login"
	password := "password"
	passwordHash := []byte{}
	publicKey := []byte{}
	privateKeyCipher := []byte{}

	user := &domain.User{
		ID:               "1",
		Login:            login,
		PasswordHash:     passwordHash,
		PublicKey:        publicKey,
		PrivateKeyCipher: privateKeyCipher,
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockUsecaseInterface(ctrl)
	mockUsecase.EXPECT().Register(
		gomock.Any(),
		login,
		password,
		publicKey,
		privateKeyCipher,
	).Return(
		user,
		nil,
	)

	type want struct {
		userID string
		err    error
	}

	tests := []struct {
		name string
		in   *pb.RegisterRequest
		want want
	}{
		{
			name: "ok",
			in: &pb.RegisterRequest{
				Login:            login,
				Password:         password,
				PublicKey:        publicKey,
				PrivateKeyCipher: privateKeyCipher,
			},
			want: want{
				userID: user.ID,
				err:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcHandler := &GRPCHandler{
				usecase:  mockUsecase,
				tokenKey: "",
				tokenExp: time.Minute,
				grpcMethodsForAuthenticateUnaryInterceptor: nil,
			}

			out, err := grpcHandler.Register(context.Background(), tt.in)
			if tt.want.err != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			accessToken := out.AccessToken

			userID, err := grpcHandler.getUserID(accessToken)
			require.NoError(t, err)

			assert.Equal(t, tt.want.userID, userID)
		})
	}
}

func TestGRPCHandler_Login(t *testing.T) {
	userID := "1"
	login := "login"
	password := "password"
	passwordHash := []byte{1, 2, 3}
	publicKey := []byte{4, 5, 6}
	privateKeyCipher := []byte{7, 8, 9}
	internalErr := errors.New("internal error")

	user := &domain.User{
		ID:               userID,
		Login:            login,
		PasswordHash:     passwordHash,
		PublicKey:        publicKey,
		PrivateKeyCipher: privateKeyCipher,
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type want struct {
		publicKey        []byte
		privateKeyCipher []byte
		userID           string
		err              error
	}

	tests := []struct {
		name            string
		mockUsecaseFunc func() *mocks.MockUsecaseInterface
		in              *pb.LoginRequest
		want            want
	}{
		{
			name: "ok",
			mockUsecaseFunc: func() *mocks.MockUsecaseInterface {
				mockUsecase := mocks.NewMockUsecaseInterface(ctrl)
				mockUsecase.EXPECT().Login(
					gomock.Any(),
					login,
					password,
				).Return(
					user,
					nil,
				)

				return mockUsecase
			},
			in: &pb.LoginRequest{
				Login:    login,
				Password: password,
			},
			want: want{
				publicKey:        publicKey,
				privateKeyCipher: privateKeyCipher,
				userID:           userID,
				err:              nil,
			},
		},
		{
			name: "internal error",
			mockUsecaseFunc: func() *mocks.MockUsecaseInterface {
				mockUsecase := mocks.NewMockUsecaseInterface(ctrl)
				mockUsecase.EXPECT().Login(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(
					nil,
					internalErr,
				)

				return mockUsecase
			},
			in: &pb.LoginRequest{
				Login:    login,
				Password: password,
			},
			want: want{
				publicKey:        nil,
				privateKeyCipher: nil,
				userID:           "",
				err:              status.Error(codes.Internal, InternalErrorMessage),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grpcHandler := &GRPCHandler{
				usecase:  tt.mockUsecaseFunc(),
				tokenKey: "123",
				tokenExp: time.Minute,
				grpcMethodsForAuthenticateUnaryInterceptor: map[string]struct{}{},
			}

			out, err := grpcHandler.Login(context.Background(), tt.in)
			if tt.want.err != nil {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			assert.Equal(t, tt.want.publicKey, out.PublicKey)
			assert.Equal(t, tt.want.privateKeyCipher, out.PrivateKeyCipher)

			accessToken := out.AccessToken

			actualUserID, err := grpcHandler.getUserID(accessToken)
			require.NoError(t, err)

			assert.Equal(t, tt.want.userID, actualUserID)
		})
	}
}
