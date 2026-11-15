// Command migrate applies pending migrations to the database in DATABASE_URL.
// It doesn't read .env, so it only touches the database it is given. make
// schema runs it against a scratch database.
package main

import (
	"believer/movies/db"
	"context"
	"log"
)

func main() {
	if err := db.InitializeConnection(); err != nil {
		log.Fatal(err)
	}

	defer db.CloseConnection()

	if err := db.Migrate(context.Background(), db.Client.DB); err != nil {
		log.Fatal(err)
	}
}
