package gRPC

import (
	gRPC2 "Demonstration-Service/api/gRPC"
	"Demonstration-Service/internal/Application/Contracts/OrdersServices"
	"Demonstration-Service/internal/Presentation/Servers/gRPC"
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
	gRPC2.RegisterOrderServiceServer(grpcServer, orderService)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	log.Println("gRPC server is listening: 50051")
}
