package convert

import (
	"Demonstration-Service/api/grpcAPI"
	"Demonstration-Service/internal/Application/Domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func OrderToPb(order Domain.Order) grpcAPI.Order {
	pbItems := make([]*grpcAPI.Item, len(order.Items))
	for i, item := range order.Items {
		pbItems[i] = &grpcAPI.Item{
			ChrtId:      int32(item.ChrtID),
			TrackNumber: item.TrackNumber,
			Price:       int32(item.Price),
			Rid:         item.Rid,
			Name:        item.Name,
			Sale:        int32(item.Sale),
			Size:        item.Size,
			TotalPrice:  int32(item.TotalPrice),
			NmId:        int32(item.NmID),
			Brand:       item.Brand,
			Status:      int32(item.Status),
		}
	}
	return grpcAPI.Order{
		OrderUid:    order.OrderUID,
		TrackNumber: order.TrackNumber,
		Entry:       order.Entry,
		Delivery: &grpcAPI.Delivery{
			Name:    order.Delivery.Name,
			Phone:   order.Delivery.Phone,
			Zip:     order.Delivery.Zip,
			City:    order.Delivery.City,
			Address: order.Delivery.Address,
			Region:  order.Delivery.Region,
			Email:   order.Delivery.Email,
		},
		Payment: &grpcAPI.Payment{
			Transaction:  order.Payment.Transaction,
			RequestId:    order.Payment.RequestID,
			Currency:     order.Payment.Currency,
			Provider:     order.Payment.Provider,
			Amount:       int32(order.Payment.Amount),
			PaymentDt:    int32(order.Payment.PaymentDT),
			Bank:         order.Payment.Bank,
			DeliveryCost: int32(order.Payment.DeliveryCost),
			GoodsTotal:   int32(order.Payment.GoodsTotal),
			CustomFee:    int32(order.Payment.CustomFee),
		},
		Items:             pbItems,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerId:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		Shardkey:          order.Shardkey,
		SmId:              int32(order.SmID),
		DateCreated:       timestamppb.New(order.DateCreated),
		OofShard:          order.OofShard,
	}
}
