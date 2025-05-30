package db

import (
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/qustavo/dotsql"
	"github.com/swithek/dotsqlx"
)

var (
	Client *sqlx.DB
	Dot    *dotsqlx.DotSqlx
)

func InitializeConnection() error {
	connectionString := os.Getenv("DATABASE_URL")
	db := sqlx.MustConnect("postgres", connectionString)

	if err := db.Ping(); err != nil {
		return err
	}

	log.Println("Connected to database")

	files := []string{
		"./db/awardQueries.sql",
		"./db/genreQueries.sql",
		"./db/languageQueries.sql",
		"./db/movieQueries.sql",
		"./db/queries.sql",
		"./db/ratingQueries.sql",
		"./db/seriesQueries.sql",
		"./db/statsQueries.sql",
		"./db/watchlistQueries.sql",
	}

	var queries []*dotsql.DotSql

	// Load all query files
	for _, file := range files {
		q, err := dotsql.LoadFromFile(file)

		if err != nil {
			return err
		}

		queries = append(queries, q)
	}

	dot := dotsql.Merge(queries...)
	dotx := dotsqlx.Wrap(dot)

	// Set the global DBClient variable to the db connection
	Client = db
	Dot = dotx

	return nil
}

func CloseConnection() {
	err := Client.Close()

	if err != nil {
		log.Fatal("Failed to close connection to database")
	}
}
