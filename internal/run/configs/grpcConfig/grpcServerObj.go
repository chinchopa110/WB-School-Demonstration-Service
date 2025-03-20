package grpcConfig

import (
	"Demonstration-Service/api/grpcAPI"
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/gRPC"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	server *grpc.Server
	logger *zap.Logger
}

func NewServer(service OrdersServices.IGetService, logger *zap.Logger) *Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(gRPC.ChainUnaryServer(
			gRPC.PanicRecoveryInterceptor(logger),
			gRPC.LoggingInterceptor(logger),
		)),
	)

	orderService := gRPC.NewServer(service)
	grpcAPI.RegisterOrderServiceServer(grpcServer, orderService)

	return &Server{
		server: grpcServer,
		logger: logger,
	}
}

func (s *Server) Start(address string) error {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s.logger.Info("gRPC server is starting", zap.String("address", address))
	return s.server.Serve(lis)
}

func (s *Server) Stop() {
	s.server.GracefulStop()
	s.logger.Info("gRPC server stopped")
}
