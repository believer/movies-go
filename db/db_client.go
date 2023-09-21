package db

import (
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var Client *sqlx.DB

func InitializeConnection() {
	connectionString := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", connectionString)

	if err != nil {
		panic(err)
	}

	// Check if connection is alive
	err = db.Ping()

	if err != nil {
		panic(err)
	}

	// Set the global DBClient variable to the db connection
	Client = db
}
