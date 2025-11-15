package db

import (
	"os"

	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	Client *sqlx.DB
)

func InitializeConnection() error {
	connectionString := os.Getenv("DATABASE_URL")
	db := sqlx.MustConnect("postgres", connectionString)

	if err := db.Ping(); err != nil {
		return err
	}

	log.Info("Connected to database")

	// Set the global DBClient variable to the db connection
	Client = db

	return nil
}

func CloseConnection() {
	err := Client.Close()

	if err != nil {
		log.Error("Failed to close connection to database")
	}
}
