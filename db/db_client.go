package db

import (
	"log/slog"
	"os"

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

	slog.Info("Connected to database")

	// Set the global DBClient variable to the db connection
	Client = db

	return nil
}

func CloseConnection() {
	err := Client.Close()

	if err != nil {
		slog.Error("Failed to close connection to database")
	}
}
