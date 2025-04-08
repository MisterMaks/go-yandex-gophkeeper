package logger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Log is global logger.
var Log *zap.Logger = zap.NewNop()

// LoggerKeyType is type for LoggerKey constant.
type LoggerKeyType string

// Constants for logger.
const (
	MethodKey            string        = "method"
	URIKey               string        = "uri"
	RequestIDKey         string        = "request_id"
	ExecutionDurationKey string        = "execution_duration"
	StatusCodeKey        string        = "status_code"
	StatusKey            string        = "status"
	ResponseBodySizeBKey string        = "response_body_size_B"
	LoggerKey            LoggerKeyType = "logger_key"
)

// New init logger.
func New(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

func generateRequestID() string {
	id := uuid.New()
	return id.String()
}

// RequestLoggerUnaryInterceptor is logger unary interceptor for grpc handlers.
func RequestLoggerUnaryInterceptor(ctx context.Context, req any, si *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	requestID := generateRequestID()
	Log.Info("got incoming GRPC request",
		zap.String(MethodKey, "POST"),
		zap.Any(URIKey, si.FullMethod),
		zap.String(RequestIDKey, requestID),
	)
	ctxLogger := Log.With(
		zap.String(RequestIDKey, requestID),
	)
	ctx = context.WithValue(ctx, LoggerKey, ctxLogger)
	now := time.Now()
	m, err := handler(ctx, req)

	responseSize := 0
	if m != nil {
		responseSize = proto.Size(m.(proto.Message))
	}

	Log.Info("processed incoming GRPC request",
		zap.Any(StatusCodeKey, uint32(status.Code(err))),
		zap.Any(StatusKey, status.Code(err)),
		zap.Int(ResponseBodySizeBKey, responseSize),
		zap.Duration(ExecutionDurationKey, time.Since(now)),
		zap.String(RequestIDKey, requestID),
	)
	return m, err
}

// GetContextLogger gets logger from context.
func GetContextLogger(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Log
	}
	logger, ok := ctx.Value(LoggerKey).(*zap.Logger)
	if !ok || logger == nil {
		return Log
	}
	return logger
}
