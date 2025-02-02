package gRPC

import (
	"context"
	"runtime/debug"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		code := status.Code(err)
		logger.Info("grpcConfig request",
			zap.String("method", info.FullMethod),
			zap.String("code", code.String()),
			zap.Duration("duration", duration),
			zap.Any("request", req),
		)

		if err != nil {
			logger.Error("grpcConfig request failed",
				zap.String("method", info.FullMethod),
				zap.Error(err),
				zap.Any("request", req),
			)
		}
		return resp, err
	}
}

func PanicRecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = status.Errorf(codes.Internal, "grpcConfig server panic: %v", r)
				logger.Error("grpcConfig panic recovery",
					zap.Any("error", r),
					zap.String("method", info.FullMethod),
					zap.String("stack", string(debug.Stack())),
				)
			}
		}()
		return handler(ctx, req)
	}
}

func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		chain := func(current grpc.UnaryServerInterceptor, currentHandler grpc.UnaryHandler) grpc.UnaryHandler {
			return func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return current(currentCtx, currentReq, info, currentHandler)
			}
		}

		chainedHandler := handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			chainedHandler = chain(interceptors[i], chainedHandler)
		}

		return chainedHandler(ctx, req)
	}
}
