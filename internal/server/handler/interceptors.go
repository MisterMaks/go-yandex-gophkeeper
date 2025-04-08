package handler

import (
	"context"
	"strings"

	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/logger"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Claims is jwt.RegisteredClaims with UserID field.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func (h *GRPCHandler) getUserID(tokenString string) (string, error) {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.tokenKey), nil
	})
	if err != nil {
		return "", err
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, nil
}

// AuthenticateUnaryInterceptor is unary interceptor for auths user.
func (h *GRPCHandler) AuthenticateUnaryInterceptor(ctx context.Context, req any, si *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if _, ok := h.grpcMethodsForAuthenticateUnaryInterceptor[si.FullMethod]; !ok {
		return handler(ctx, req)
	}

	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get(AuthorizationHeaderKey)
		if len(values) > 0 {
			token = strings.TrimPrefix(values[0], BearerKey)
		}
	}

	if len(token) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "missing %s metadata", AuthorizationHeaderKey)
	}

	userID, err := h.getUserID(token)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	ctx = context.WithValue(ctx, UserIDKey, userID)

	ctxLogger := logger.GetContextLogger(ctx)
	ctxLogger = ctxLogger.With(zap.String(string(UserIDKey), userID))
	ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

	return handler(ctx, req)
}
