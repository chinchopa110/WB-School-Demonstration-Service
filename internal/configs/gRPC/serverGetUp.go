package gRPC

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/gRPC"
	"Demonstration-Service/internal/Presentation/Servers/gRPC/api"
	"google.golang.org/grpc"
	"log"
	"net"
)

func ServerGetUp(service OrdersServices.IGetService) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	orderService := gRPC.NewServer(service)
	api.RegisterOrderServiceServer(grpcServer, orderService)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	log.Println("gRPC server is listening: 50051")
}
