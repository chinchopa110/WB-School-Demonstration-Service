package gRPC

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/gRPC/api"
	"Demonstration-Service/internal/Presentation/Servers/gRPC/convert"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

type Server struct {
	api.UnimplementedOrderServiceServer
	service OrdersServices.IGetService
}

func NewServer(service OrdersServices.IGetService) *Server {
	return &Server{service: service}
}

func (s *Server) GetOrder(ctx context.Context, req *api.GetOrderRequest) (*api.GetOrderResponse, error) {
	orderId := req.GetId()
	if orderId == "" {
		return nil, status.Error(codes.InvalidArgument, "Order ID cannot be empty")
	}

	intOrderId, err := strconv.Atoi(orderId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid Order ID, must be an integer")
	}

	order, err := s.service.GetById(intOrderId)

	if err != nil {
		return nil, status.Errorf(codes.NotFound, "order with id: %s, not found: %v", orderId, err)
	}
	pbOrder := convert.OrderToPb(order)

	return &api.GetOrderResponse{Order: &pbOrder}, nil
}
