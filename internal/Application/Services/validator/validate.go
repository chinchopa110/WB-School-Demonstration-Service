package validator

import (
	"Demonstration-Service/internal/Application/Domain"
	"errors"
	"fmt"
)

func ValidateOrder(order Domain.Order) error {
	if order.OrderUID == "" {
		return errors.New("order_uid is required")
	}
	if order.TrackNumber == "" {
		return errors.New("track_number is required")
	}
	if order.Entry == "" {
		return errors.New("entry is required")
	}
	if order.Locale == "" {
		return errors.New("locale is required")
	}
	if order.CustomerID == "" {
		return errors.New("customer_id is required")
	}
	if order.DeliveryService == "" {
		return errors.New("delivery_service is required")
	}
	if order.Shardkey == "" {
		return errors.New("shardkey is required")
	}
	if order.SmID <= 0 {
		return errors.New("sm_id must be greater than zero")
	}
	if order.DateCreated.IsZero() {
		return errors.New("date_created is required")
	}
	if order.OofShard == "" {
		return errors.New("oof_shard is required")
	}
	if err := validateDelivery(order.Delivery); err != nil {
		return fmt.Errorf("delivery validation error: %w", err)
	}

	if err := validatePayment(order.Payment); err != nil {
		return fmt.Errorf("payment validation error: %w", err)
	}

	if err := validateItems(order.Items); err != nil {
		return fmt.Errorf("items validation error: %w", err)
	}
	return nil
}

func validateDelivery(delivery Domain.Delivery) error {
	if delivery.Name == "" {
		return errors.New("delivery name is required")
	}
	if delivery.Phone == "" {
		return errors.New("delivery phone is required")
	}
	if delivery.Zip == "" {
		return errors.New("delivery zip is required")
	}
	if delivery.City == "" {
		return errors.New("delivery city is required")
	}
	if delivery.Address == "" {
		return errors.New("delivery address is required")
	}
	if delivery.Region == "" {
		return errors.New("delivery region is required")
	}
	if delivery.Email == "" {
		return errors.New("delivery email is required")
	}
	return nil
}

func validatePayment(payment Domain.Payment) error {
	if payment.Transaction == "" {
		return errors.New("payment transaction is required")
	}
	if payment.Currency == "" {
		return errors.New("payment currency is required")
	}
	if payment.Provider == "" {
		return errors.New("payment provider is required")
	}
	if payment.Amount <= 0 {
		return errors.New("payment amount must be greater than zero")
	}
	if payment.PaymentDT <= 0 {
		return errors.New("payment payment_dt must be greater than zero")
	}
	if payment.Bank == "" {
		return errors.New("payment bank is required")
	}
	return nil
}

func validateItems(items []Domain.Item) error {
	if len(items) == 0 {
		return errors.New("at least one item is required")
	}
	for _, item := range items {
		if item.ChrtID <= 0 {
			return errors.New("item chrt_id must be greater than zero")
		}
		if item.TrackNumber == "" {
			return errors.New("item track_number is required")
		}
		if item.Price <= 0 {
			return errors.New("item price must be greater than zero")
		}
		if item.Rid == "" {
			return errors.New("item rid is required")
		}
		if item.Name == "" {
			return errors.New("item name is required")
		}
		if item.Size == "" {
			return errors.New("item size is required")
		}
		if item.TotalPrice <= 0 {
			return errors.New("item total_price must be greater than zero")
		}
		if item.NmID <= 0 {
			return errors.New("item nm_id must be greater than zero")
		}
		if item.Brand == "" {
			return errors.New("item brand is required")
		}
		if item.Status <= 0 {
			return errors.New("item status must be greater than zero")
		}
	}
	return nil
}
