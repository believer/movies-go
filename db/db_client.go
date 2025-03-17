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
	err := db.Ping()

	if err != nil {
		return err
	} else {
		log.Println("Connected to database")
	}

	generalQueries, err := dotsql.LoadFromFile("./db/queries.sql")

	if err != nil {
		return err
	}

	seriesQueries, err := dotsql.LoadFromFile("./db/seriesQueries.sql")

	if err != nil {
		return err
	}

	watchlistQueries, err := dotsql.LoadFromFile("./db/watchlistQueries.sql")

	if err != nil {
		return err
	}

	statsQueries, err := dotsql.LoadFromFile("./db/statsQueries.sql")

	if err != nil {
		return err
	}

	genreQueries, err := dotsql.LoadFromFile("./db/genreQueries.sql")

	if err != nil {
		return err
	}

	movieQueries, err := dotsql.LoadFromFile("./db/movieQueries.sql")

	if err != nil {
		return err
	}

	languageQueries, err := dotsql.LoadFromFile("./db/languageQueries.sql")

	if err != nil {
		return err
	}

	awardQueries, err := dotsql.LoadFromFile("./db/awardQueries.sql")

	if err != nil {
		return err
	}

	dot := dotsql.Merge(
		generalQueries,
		genreQueries,
		movieQueries,
		seriesQueries,
		statsQueries,
		watchlistQueries,
		languageQueries,
		awardQueries,
	)

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
