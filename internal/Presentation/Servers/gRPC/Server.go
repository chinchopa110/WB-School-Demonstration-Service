package gRPC

import (
	"Demonstration-Service/api/grpcAPI"
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/gRPC/convert"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	grpcAPI.UnimplementedOrderServiceServer
	service OrdersServices.IGetService
}

func NewServer(service OrdersServices.IGetService) *Server {
	return &Server{service: service}
}

func (s *Server) GetOrder(ctx context.Context, req *grpcAPI.GetOrderRequest) (*grpcAPI.GetOrderResponse, error) {
	orderId := req.GetId()
	if orderId == "" {
		return nil, status.Error(codes.InvalidArgument, "Order ID cannot be empty")
	}

	order, err := s.service.GetById(orderId)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order with id: %s, not found: %v", orderId, err)
	}
	pbOrder := convert.OrderToPb(order)

	return &grpcAPI.GetOrderResponse{Order: &pbOrder}, nil
}
