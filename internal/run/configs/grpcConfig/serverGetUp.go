package grpcConfig

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"

	"go.uber.org/zap"
)

func ServerGetUp(service OrdersServices.IGetService, logger *zap.Logger) *Server {
	return NewServer(service, logger)
}
