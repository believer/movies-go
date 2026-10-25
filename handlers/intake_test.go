package handlers

import (
	"believer/movies/db/dbtest"
	"believer/movies/services/intake"
	"believer/movies/services/tmdb/tmdbtest"
	"fmt"
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
	dbtest.Truncate(t, db, "movie", "person", "genre", "language", "production_company", "series", "user", "watchlist", "now_playing")

	// Awards are read from the CSV files at the repository root
	t.Chdir("..")

	fake := &tmdbtest.Fake{}

	return testIntake{intake.New(db, fake), fake, db}
}

func (ti testIntake) seedUser(t *testing.T, id int) {
	t.Helper()

	_, err := ti.db.Exec(`INSERT INTO "user" (id, username, password_hash, name) VALUES ($1, $2, '', $2)`, id, fmt.Sprintf("user%d", id))

	if err != nil {
		t.Fatal(err)
	}
}

func (ti testIntake) count(t *testing.T, query string, args ...any) int {
	t.Helper()

	var n int

	if err := ti.db.Get(&n, query, args...); err != nil {
		t.Fatal(err)
	}

	return n
}
