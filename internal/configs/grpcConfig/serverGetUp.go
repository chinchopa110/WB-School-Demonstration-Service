package grpcConfig

import (
	"Demonstration-Service/api/grpcAPI"
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/gRPC"
	"Demonstration-Service/internal/configs"
	"log"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func ServerGetUp(service OrdersServices.IGetService) {
	logger, err := configs.InitLogger("logs/grpc.log")
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatal("failed to listen: %v", zap.Error(err))
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			log.Fatalf("failed to sync logger: %v", err)
		}
	}()
	logger.Info("gRPC server is starting", zap.String("address", ":50051"))

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(gRPC.ChainUnaryServer(
			gRPC.PanicRecoveryInterceptor(logger),
			gRPC.LoggingInterceptor(logger),
		)),
	)

	orderService := gRPC.NewServer(service)
	grpcAPI.RegisterOrderServiceServer(grpcServer, orderService)

	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatal("failed to serve gRPC server", zap.Error(err))
	}
	logger.Info("gRPC server is listening", zap.String("address", ":50051"))
}
