package configs

import (
	"database/sql"
	_ "github.com/lib/pq"
)

func GetUpSQL() (*sql.DB, error) {
	connStr := "user=postgres password=123 dbname=wb1 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}
