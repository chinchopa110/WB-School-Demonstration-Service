package configs

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func GetUpSQL() *sql.DB {
	connStr := "user=postgres password=123 dbname=wb1 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Could not connect to the database: %s\n", err)
	}

	return db
}
