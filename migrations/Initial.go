package migrations

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func RunMigrations(db *sql.DB) {

	_, err := db.Exec(`
    -- Таблица заказов
        CREATE TABLE IF NOT EXISTS orders (
            order_uid TEXT PRIMARY KEY,
            track_number TEXT,
            entry TEXT,
            locale TEXT,
            internal_signature TEXT,
            customer_id TEXT,
            delivery_service TEXT,
            shardkey TEXT,
            sm_id INTEGER,
            date_created TIMESTAMP WITH TIME ZONE,
            oof_shard TEXT
        );

        -- Таблица доставок
        CREATE TABLE IF NOT EXISTS deliveries (
            order_uid TEXT PRIMARY KEY REFERENCES orders(order_uid),
            name TEXT,
            phone TEXT,
            zip TEXT,
            city TEXT,
            address TEXT,
            region TEXT,
            email TEXT
        );

        -- Таблица платежей
        CREATE TABLE IF NOT EXISTS payments (
            transaction TEXT PRIMARY KEY,
            order_uid TEXT REFERENCES orders(order_uid),
            request_id TEXT,
            currency TEXT,
            provider TEXT,
            amount INTEGER,
            payment_dt INTEGER,
            bank TEXT,
            delivery_cost INTEGER,
            goods_total INTEGER,
            custom_fee INTEGER
        );

        -- Таблица товаров
        CREATE TABLE IF NOT EXISTS items (
            chrt_id INTEGER ,
            order_uid TEXT REFERENCES orders(order_uid),
            track_number TEXT,
            price INTEGER,
            rid TEXT,
            name TEXT,
            sale INTEGER,
            size TEXT,
            total_price INTEGER,
            nm_id INTEGER,
            brand TEXT,
            status INTEGER,
            PRIMARY KEY(chrt_id, order_uid)
            
        -- Inbox
		CREATE TABLE IF NOT EXISTS inbox (
			message_id VARCHAR(255) PRIMARY KEY,
			message_type VARCHAR(255) NOT NULL,
			payload JSONB NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			processed_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			attempts INT DEFAULT 0,
			error_message TEXT
		);
		
		CREATE INDEX IF NOT EXISTS idx_inbox_status ON inbox(status);
        );
    `)
	if err != nil {
		log.Fatalf("Ошибка при выполнении миграции: %v", err)
	}

	fmt.Println("Миграция успешно применена.")
}
