package main

import (
	"Demonstration-Service/api/gRPC"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v, status: %v", err, status.Convert(err).Code())
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}()

	client := gRPC.NewOrderServiceClient(conn)
	orderId := "123"
	req := &gRPC.GetOrderRequest{
		Id: orderId,
	}

	resp, err := client.GetOrder(ctx, req)

	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.NotFound {
				fmt.Printf("Order not found with id: %v", orderId)
				return
			}

			if st.Code() == codes.InvalidArgument {
				fmt.Printf("Invalid order id: %v", orderId)
				return
			}
		}

		log.Fatalf("Could not get order: %v", err)
		return
	}

	if resp != nil {
		order := resp.GetOrder()

		fmt.Println("Order ID: ", order.GetOrderUid())

		fmt.Println("TrackNumber: ", order.GetTrackNumber())
		fmt.Println("Entry: ", order.GetEntry())
		fmt.Println("Locale: ", order.GetLocale())
		fmt.Println("InternalSignature:", order.GetInternalSignature())
		fmt.Println("CustomerId:", order.GetCustomerId())
		fmt.Println("DeliveryService:", order.GetDeliveryService())
		fmt.Println("Shardkey:", order.GetShardkey())
		fmt.Println("SmId:", order.GetSmId())
		fmt.Println("DateCreated:", order.GetDateCreated())
		fmt.Println("OofShard:", order.GetOofShard())

		if delivery := order.GetDelivery(); delivery != nil {
			fmt.Println("Delivery:")
			fmt.Println("\tName: ", delivery.GetName())
			fmt.Println("\tPhone:", delivery.GetPhone())
			fmt.Println("\tZip:", delivery.GetZip())
			fmt.Println("\tCity:", delivery.GetCity())
			fmt.Println("\tAddress:", delivery.GetAddress())
			fmt.Println("\tRegion:", delivery.GetRegion())
			fmt.Println("\tEmail:", delivery.GetEmail())
		} else {
			fmt.Println("Delivery: nil")
		}

		if payment := order.GetPayment(); payment != nil {
			fmt.Println("Payment:")
			fmt.Println("\tTransaction:", payment.GetTransaction())
			fmt.Println("\tRequestId:", payment.GetRequestId())
			fmt.Println("\tCurrency:", payment.GetCurrency())
			fmt.Println("\tProvider:", payment.GetProvider())
			fmt.Println("\tAmount:", payment.GetAmount())
			fmt.Println("\tPaymentDt:", payment.GetPaymentDt())
			fmt.Println("\tDeliveryCost:", payment.GetDeliveryCost())
			fmt.Println("\tCustomFee:", payment.GetCustomFee())
			fmt.Println("\tBank:", payment.GetBank())
			fmt.Println("\tGoodsTotal:", payment.GetGoodsTotal())
		} else {
			fmt.Println("Payment: nil")
		}

		fmt.Println("Items:")
		if items := order.GetItems(); items != nil {
			for _, item := range items {
				fmt.Println("\t Item ChrtId:", item.GetChrtId())
				fmt.Println("\t Item Price:", item.GetPrice())
				fmt.Println("\t Item TrackNumber:", item.GetTrackNumber())
				fmt.Println("\t Item Rid:", item.GetRid())
				fmt.Println("\t Item Name:", item.GetName())
				fmt.Println("\t Item Sale:", item.GetSale())
				fmt.Println("\t Item Size:", item.GetSize())
				fmt.Println("\t Item TotalPrice:", item.GetTotalPrice())
				fmt.Println("\t Item NmId:", item.GetNmId())
				fmt.Println("\t Item Brand:", item.GetBrand())
				fmt.Println("\t Item Status:", item.GetStatus())
			}
		} else {
			fmt.Println("Items: nil")
		}

	} else {
		fmt.Println("Empty response")
	}
}
