package dataAccess

import (
	"Demonstration-Service/internal/Application/Domain"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"time"
)

type OrdersRepo struct {
	db *sql.DB
}

func NewOrdersRepo(db *sql.DB) *OrdersRepo {
	return &OrdersRepo{db: db}
}

func (repo *OrdersRepo) IsExist(id string) bool {
	var exists bool
	err := repo.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid = $1)", id).Scan(&exists)

	if err != nil {
		return false
	}
	return exists
}

func (repo *OrdersRepo) Read(id string) (Domain.Order, error) {
	order := Domain.Order{}
	var delivery Domain.Delivery
	var payment Domain.Payment
	var items []Domain.Item

	err := repo.db.QueryRow(`
        SELECT 
            order_uid, 
            track_number, 
            entry, 
            locale, 
            internal_signature, 
            customer_id, 
            delivery_service,
            shardkey, 
            sm_id, 
            date_created, 
            oof_shard
        FROM orders
        WHERE order_uid = $1
    `, id).Scan(
		&order.OrderUID,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerID,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmID,
		&order.DateCreated,
		&order.OofShard,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Domain.Order{}, fmt.Errorf("order with id %s not found: %w", id, err)
		}
		return Domain.Order{}, fmt.Errorf("error reading order data: %w", err)
	}

	err = repo.db.QueryRow(`
        SELECT 
            name, 
            phone, 
            zip, 
            city, 
            address, 
            region, 
            email
        FROM deliveries
        WHERE order_uid = $1
    `, id).Scan(
		&delivery.Name,
		&delivery.Phone,
		&delivery.Zip,
		&delivery.City,
		&delivery.Address,
		&delivery.Region,
		&delivery.Email,
	)
	if err != nil {
		return Domain.Order{}, fmt.Errorf("error reading delivery data: %w", err)
	}
	delivery.OrderUID = order.OrderUID
	order.Delivery = delivery

	err = repo.db.QueryRow(`
        SELECT 
            _transaction, 
            request_id, 
            currency, 
            provider, 
            amount, 
            payment_dt, 
            bank,
            delivery_cost, 
            goods_total, 
            custom_fee
        FROM payments
        WHERE order_uid = $1
    `, id).Scan(
		&payment.Transaction,
		&payment.RequestID,
		&payment.Currency,
		&payment.Provider,
		&payment.Amount,
		&payment.PaymentDT,
		&payment.Bank,
		&payment.DeliveryCost,
		&payment.GoodsTotal,
		&payment.CustomFee,
	)
	if err != nil {
		return Domain.Order{}, fmt.Errorf("error reading payment data: %w", err)
	}
	payment.OrderUID = order.OrderUID
	order.Payment = payment

	rows, err := repo.db.Query(`
        SELECT
            order_uid,
            chrt_id, 
            track_number, 
            price, 
            rid, 
            name, 
            sale, 
            _size, 
            total_price, 
            nm_id, 
            brand, 
            status
        FROM items
        WHERE order_uid = $1
    `, id)
	if err != nil {
		return Domain.Order{}, fmt.Errorf("error reading items data: %w", err)
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			log.Printf("error closing rows in items table: %v", closeErr)
		}
	}()

	for rows.Next() {
		item := Domain.Item{}
		err := rows.Scan(
			&item.OrderUID,
			&item.ChrtID,
			&item.TrackNumber,
			&item.Price,
			&item.Rid,
			&item.Name,
			&item.Sale,
			&item.Size,
			&item.TotalPrice,
			&item.NmID,
			&item.Brand,
			&item.Status,
		)
		if err != nil {
			return Domain.Order{}, fmt.Errorf("error scanning item data: %w", err)
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return Domain.Order{}, fmt.Errorf("error iterating through item rows: %w", err)
	}

	order.Items = items
	return order, nil
}

func (repo *OrdersRepo) Save(order Domain.Order, ctx context.Context) error {
	message := ctx.Value("message").(kafka.Message)
	messageID := message.Key
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	var existingStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT status FROM inbox WHERE message_id = $1 FOR UPDATE
	`, messageID).Scan(&existingStatus)

	if err == nil && existingStatus == "processed" {
		tx.Rollback()
		return nil
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		tx.Rollback()
		return fmt.Errorf("error checking inbox: %w", err)
	}

	payload, err := json.Marshal(order)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error marshaling order: %w", err)
	}

	if errors.Is(err, sql.ErrNoRows) {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO inbox (message_id, message_type, payload, status)
			VALUES ($1, $2, $3, 'processing')
		`, messageID, "order_created", payload)
	} else {
		_, err = tx.ExecContext(ctx, `
			UPDATE inbox 
			SET status = 'processing', 
				attempts = attempts + 1,
				error_message = NULL
			WHERE message_id = $1
		`, messageID)
	}

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error updating inbox: %w", err)
	}

	if err := repo.saveOrderTx(tx, order); err != nil {
		_, rbErr := tx.ExecContext(ctx, `
			UPDATE inbox 
			SET status = 'failed', 
				error_message = $1
			WHERE message_id = $2
		`, err.Error(), messageID)
		if rbErr != nil {
			log.Printf("failed to update inbox status: %v", rbErr)
		}
		tx.Rollback()
		return fmt.Errorf("error saving order: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE inbox 
		SET status = 'processed',
			processed_at = NOW()
		WHERE message_id = $1
	`, messageID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error marking message as processed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	if reader, ok := ctx.Value("message reader").(*kafka.Reader); ok {
		if message, ok := ctx.Value("message").(kafka.Message); ok {
			if err := reader.CommitMessages(ctx, message); err != nil {
				tx.Rollback()
				return fmt.Errorf("error committing kafka message: %w", err)
			}
		}
	}

	return nil
}

func (repo *OrdersRepo) saveOrderTx(tx *sql.Tx, order Domain.Order) error {
	_, err := tx.Exec(`
        INSERT INTO orders ( order_uid,
                            track_number,
                            entry,
                            locale,
                            internal_signature,
                            customer_id, 
                            delivery_service,
                            shardkey,
                            sm_id,
                            date_created,
                            oof_shard
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `,
		order.OrderUID,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerID,
		order.DeliveryService,
		order.Shardkey,
		order.SmID,
		order.DateCreated,
		order.OofShard)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error inserting into orders table: %w for order_uid: %s", err, order.OrderUID)
	}

	_, err = tx.Exec(`
        INSERT INTO deliveries ( order_uid,
                                name,
                                phone,
                                zip,
                                city,
                                address,
                                region,
                                email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `,
		order.OrderUID,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error inserting into deliveries table: %w for order_uid: %s", err, order.OrderUID)
	}

	_, err = tx.Exec(`
        INSERT INTO payments ( _transaction,
                              order_uid,
                              request_id,
                              currency,
                              provider,
                              amount,
                              payment_dt,
                              bank,
                              delivery_cost,
                              goods_total,
                              custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `,
		order.Payment.Transaction,
		order.OrderUID,
		order.Payment.RequestID,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDT,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("error inserting into payments table: %w for order_uid: %s", err, order.OrderUID)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO items ( chrt_id,
                               order_uid,
                               track_number,
                               price,
                               rid,
                               name,
                               sale,
                               _size, 
                               total_price,
                               nm_id,
                               brand,
                               status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        `,
			item.ChrtID,
			order.OrderUID,
			item.TrackNumber,
			item.Price,
			item.Rid,
			item.Name,
			item.Sale,
			item.Size,
			item.TotalPrice,
			item.NmID,
			item.Brand,
			item.Status)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("error inserting into items table: %w for order_uid: %s and chrt_id %d", err, order.OrderUID, item.ChrtID)
		}
	}

	return nil
}

func (repo *OrdersRepo) ProcessFailedMessages(ctx context.Context, maxAttempts int, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			repo.retryFailedMessages(ctx, maxAttempts)
		}
	}
}

func (repo *OrdersRepo) retryFailedMessages(ctx context.Context, maxAttempts int) {
	tx, err := repo.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT message_id, payload 
		FROM inbox 
		WHERE status = 'failed' 
		AND attempts < $1
		ORDER BY created_at 
		FOR UPDATE SKIP LOCKED 
		LIMIT 100
	`, maxAttempts)
	if err != nil {
		tx.Rollback()
		log.Printf("Failed to query inbox: %v", err)
		return
	}
	defer rows.Close()

	var messages []struct {
		ID      string
		Payload []byte
	}

	for rows.Next() {
		var msg struct {
			ID      string
			Payload []byte
		}
		if err := rows.Scan(&msg.ID, &msg.Payload); err != nil {
			tx.Rollback()
			log.Printf("Failed to scan inbox row: %v", err)
			return
		}
		messages = append(messages, msg)
	}

	if len(messages) == 0 {
		tx.Rollback()
		return
	}

	for _, msg := range messages {
		var order Domain.Order
		if err := json.Unmarshal(msg.Payload, &order); err != nil {
			log.Printf("Failed to unmarshal order: %v", err)
			continue
		}

		if err := repo.saveOrderTx(tx, order); err != nil {
			_, rbErr := tx.ExecContext(ctx, `
				UPDATE inbox 
				SET status = 'failed', 
					error_message = $1,
					attempts = attempts + 1
				WHERE message_id = $2
			`, err.Error(), msg.ID)
			if rbErr != nil {
				log.Printf("failed to update inbox status: %v", rbErr)
			}
			continue
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE inbox 
			SET status = 'processed',
				processed_at = NOW()
			WHERE message_id = $1
		`, msg.ID)
		if err != nil {
			log.Printf("failed to mark message as processed: %v", err)
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
	}
}
