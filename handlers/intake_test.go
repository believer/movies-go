package handlers

import (
	"believer/movies/db/dbtest"
	"believer/movies/services/intake"
	"believer/movies/services/tmdb/tmdbtest"
	"testing"

	"github.com/jmoiron/sqlx"
)

type testIntake struct {
	intake *intake.Intake
	tmdb   *tmdbtest.Fake
	db     *sqlx.DB
}

// setupIntake returns a Movie intake backed by an empty test database and a
// fake TMDB. Tests using it skip when TEST_DATABASE_URL is unset.
func setupIntake(t *testing.T) testIntake {
	t.Helper()

	db := dbtest.Open(t)
	dbtest.Truncate(t, db, "movie", "person", "genre", "language", "production_company", "series", "watchlist", "now_playing")

	// Awards are read from the CSV files at the repository root
	t.Chdir("..")

	fake := &tmdbtest.Fake{}

	return testIntake{intake.New(db, fake), fake, db}
}
