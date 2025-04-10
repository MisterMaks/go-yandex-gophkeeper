package handler

import (
	"context"
	"testing"
	"time"

	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/handler/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestGRPCHandler_AuthenticateUnaryInterceptor(t *testing.T) {
	userID := "1"
	methodName := "MethodName"

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	handler := &GRPCHandler{
		usecase:  nil,
		tokenKey: "",
		tokenExp: 10 * time.Second,
		grpcMethodsForAuthenticateUnaryInterceptor: map[string]struct{}{
			methodName: {},
		},
	}

	jwt, err := handler.buildJWTString(userID)
	require.NoError(t, err)

	handlerFunc := func(ctx context.Context, _ any) (any, error) {
		handlerActualUserID, handlerErr := getContextUserID(ctx)
		return handlerActualUserID, handlerErr
	}

	mockSTS := mocks.NewMockServerTransportStream(ctrl)
	mockSTS.EXPECT().SetHeader(gomock.Any()).AnyTimes()

	type want struct {
		userID string
		err    error
	}

	tests := []struct {
		name       string
		ctx        context.Context
		methodName string
		want       want
	}{
		{
			name: "ok",
			ctx: metadata.NewIncomingContext(
				grpc.NewContextWithServerTransportStream(context.Background(), mockSTS),
				metadata.MD{AuthorizationHeaderKey: []string{"Bearer " + jwt}},
			),
			methodName: methodName,
			want: want{
				userID: userID,
				err:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualUserID, actualErr := handler.AuthenticateUnaryInterceptor(
				tt.ctx,
				nil,
				&grpc.UnaryServerInfo{FullMethod: tt.methodName},
				handlerFunc,
			)

			if tt.want.err != nil {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tt.want.userID, actualUserID)
		})
	}
}
