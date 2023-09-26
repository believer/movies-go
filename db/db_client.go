package db

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var Client *sqlx.DB

func InitializeConnection() error {
	connectionString := os.Getenv("DATABASE_URL")

	db := sqlx.MustConnect("postgres", connectionString)
	err := db.Ping()

	if err != nil {
		return err
	} else {
		log.Println("Connected to database")
	}

	// Set the global DBClient variable to the db connection
	Client = db

	return nil
}

func CloseConnection() {
	err := Client.Close()

	if err != nil {
		log.Fatal("Failed to close connection to database")
	}
}
