package dataAccess

import (
	"Demonstration-Service/internal/Application/Domain"
	"database/sql"
	"errors"
	"fmt"
	"log"
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

func (repo *OrdersRepo) Read(id int) (Domain.Order, error) {
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
			return Domain.Order{}, fmt.Errorf("order with id %d not found: %w", id, err)
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

func (repo *OrdersRepo) Save(order Domain.Order) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	_, err = tx.Exec(`
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
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("error rolling back transaction for inserting into orders table: %v, original error %v", rollbackErr, err)
		}
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
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("error rolling back transaction for inserting into deliveries table: %v, original error %v", rollbackErr, err)
		}
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
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("error rolling back transaction for inserting into items table: %v, original error %v", rollbackErr, err)
		}
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
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("error rolling back transaction for inserting into items table: %v, original error %v", rollbackErr, err)
			}
			return fmt.Errorf("error inserting into items table: %w for order_uid: %s and chrt_id %d", err, order.OrderUID, item.ChrtID)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}
	return nil
}
