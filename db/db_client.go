package db

import (
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var Client *sqlx.DB

func InitializeConnection() {
	connectionString := os.Getenv("DATABASE_URL")

	db := sqlx.MustConnect("postgres", connectionString)
	err := db.Ping()

	if err != nil {
		panic(err)
	} else {
		println("Connected to database")
	}

	// Set the global DBClient variable to the db connection
	Client = db
}
