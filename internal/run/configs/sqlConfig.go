package configs

import (
	"database/sql"
	_ "github.com/lib/pq"
)

var db *sql.DB

func GetUpSQL() (*sql.DB, error) {
	if db != nil {
		return db, nil
	}
	
	connStr := "user=postgres password=123 dbname=wb1 sslmode=disable"

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return db, nil
}
