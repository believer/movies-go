package db

import (
	"log/slog"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	Client *sqlx.DB
)

// The pool is capped, so a burst of slow queries waits for a connection
// instead of opening more than the 512 MB machine and the database can hold.
// Connections are recycled, so a setting changed on the role, like
// statement_timeout, reaches every connection without a restart.
const (
	maxOpenConns    = 10
	maxIdleConns    = 5
	connMaxLifetime = 30 * time.Minute
)

func InitializeConnection() error {
	connectionString := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

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
