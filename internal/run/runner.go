package run

import (
	"Demonstration-Service/internal/configs"
	"log"
)

func Run() {
	db := configs.GetUpSQL()

	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Could not close database connection: %s\n", err)
		}
	}()

	log.Printf("Done...")
}
