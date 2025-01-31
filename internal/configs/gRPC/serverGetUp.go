package gRPC

import (
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/gRPC"
	"Demonstration-Service/internal/Presentation/Servers/gRPC/api"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

func serverGetUp(service OrdersServices.IGetService) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Println("server is listening")

	grpcServer := grpc.NewServer()

	orderService := gRPC.NewServer(service)
	api.RegisterOrderServiceServer(grpcServer, orderService)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
