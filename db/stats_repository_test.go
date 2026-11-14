package db

import (
	"believer/movies/db/dbtest"
	"believer/movies/types"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

type testStats struct {
	db *sqlx.DB
}

// setupStats returns a stats repository fixture backed by an empty test
// database with users 1 and 2. Tests using it skip when TEST_DATABASE_URL is
// unset.
func setupStats(t *testing.T) testStats {
	t.Helper()

	db := dbtest.Open(t)
	dbtest.Truncate(t, db, "movie", "user", "seen", "rating")

	ts := testStats{db}
	ts.seedUser(t, 1)
	ts.seedUser(t, 2)

	return ts
}

func (ts testStats) seedUser(t *testing.T, id int) {
	t.Helper()
	ts.exec(t, `INSERT INTO "user" (id, username, password_hash, name) VALUES ($1, $2, '', $2)`, id, fmt.Sprintf("user%d", id))
}

func (ts testStats) exec(t *testing.T, query string, args ...any) {
	t.Helper()

	if _, err := ts.db.Exec(query, args...); err != nil {
		t.Fatal(err)
	}
}

func (ts testStats) movie(t *testing.T, title string) int {
	t.Helper()

	var id int

	if err := ts.db.Get(&id, `INSERT INTO movie (title, imdb_id) VALUES ($1, $1) RETURNING id`, title); err != nil {
		t.Fatal(err)
	}

	return id
}

func (ts testStats) watch(t *testing.T, userID int, movieID int, date string) {
	t.Helper()
	ts.exec(t, `INSERT INTO seen (user_id, movie_id, date) VALUES ($1, $2, $3)`, userID, movieID, date)
}

func (ts testStats) rate(t *testing.T, userID int, movieID int, rating int, createdAt string) {
	t.Helper()
	ts.exec(t, `INSERT INTO rating (user_id, movie_id, rating, created_at) VALUES ($1, $2, $3, $4)`, userID, movieID, rating, createdAt)
}

// watchedAndRated adds a movie that user 1 watched once on date and rated on
// the same date.
func (ts testStats) watchedAndRated(t *testing.T, title string, rating int, date string) int {
	t.Helper()

	id := ts.movie(t, title)
	ts.watch(t, 1, id, date)
	ts.rate(t, 1, id, rating, date)

	return id
}

// bestOfTheYear returns user 1's best of 2024.
func (ts testStats) bestOfTheYear(t *testing.T) []types.ListItem {
	t.Helper()

	items, err := NewStatsRepository(ts.db).GetBestOfTheYear("1", "2024")

	if err != nil {
		t.Fatal(err)
	}

	return items
}

type ratedTitle struct {
	Name   string
	Rating int
}

func ratedTitles(items []types.ListItem) []ratedTitle {
	titles := []ratedTitle{}

	for _, item := range items {
		titles = append(titles, ratedTitle{item.Name, item.Count})
	}

	return titles
}

func TestGetBestOfTheYear(t *testing.T) {
	t.Run("returns the top 10 by rating, earliest watch first", func(t *testing.T) {
		ts := setupStats(t)

		ts.watchedAndRated(t, "Five", 5, "2024-01-01")
		ts.watchedAndRated(t, "Nine late", 9, "2024-05-01")
		ts.watchedAndRated(t, "Ten", 10, "2024-06-01")
		ts.watchedAndRated(t, "Nine early", 9, "2024-02-01")

		// Added in reverse so the order can't come from insertion
		for day := 8; day >= 1; day-- {
			ts.watchedAndRated(t, fmt.Sprintf("Seven %d", day), 7, fmt.Sprintf("2024-03-%02d", day))
		}

		assert.Equal(t, []ratedTitle{
			{"Ten", 10},
			{"Nine early", 9},
			{"Nine late", 9},
			{"Seven 1", 7},
			{"Seven 2", 7},
			{"Seven 3", 7},
			{"Seven 4", 7},
			{"Seven 5", 7},
			{"Seven 6", 7},
			{"Seven 7", 7},
		}, ratedTitles(ts.bestOfTheYear(t)))
	})

	t.Run("returns every movie when fewer than 10 are eligible", func(t *testing.T) {
		ts := setupStats(t)

		ts.watchedAndRated(t, "Six", 6, "2024-04-01")
		ts.watchedAndRated(t, "Eight", 8, "2024-07-01")

		assert.Equal(t, []ratedTitle{{"Eight", 8}, {"Six", 6}}, ratedTitles(ts.bestOfTheYear(t)))
	})

	t.Run("returns nothing when no movies are eligible", func(t *testing.T) {
		ts := setupStats(t)

		ts.watchedAndRated(t, "Last year", 9, "2023-06-01")

		assert.Empty(t, ts.bestOfTheYear(t))
	})

	t.Run("leaves out movies watched twice in the year", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.watchedAndRated(t, "Twice", 9, "2024-02-01")
		ts.watch(t, 1, id, "2024-08-01")
		ts.watchedAndRated(t, "Once", 7, "2024-03-01")

		assert.Equal(t, []ratedTitle{{"Once", 7}}, ratedTitles(ts.bestOfTheYear(t)))
	})

	t.Run("leaves out movies also watched in another year", func(t *testing.T) {
		ts := setupStats(t)

		earlier := ts.watchedAndRated(t, "Also earlier", 9, "2024-02-01")
		ts.watch(t, 1, earlier, "2023-02-01")
		later := ts.watchedAndRated(t, "Also later", 9, "2024-02-01")
		ts.watch(t, 1, later, "2025-02-01")
		ts.watchedAndRated(t, "Only this year", 7, "2024-03-01")

		assert.Equal(t, []ratedTitle{{"Only this year", 7}}, ratedTitles(ts.bestOfTheYear(t)))
	})

	t.Run("ignores other users' watches and ratings", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.watchedAndRated(t, "Shared", 8, "2024-02-01")
		ts.watch(t, 2, id, "2023-02-01")
		ts.watch(t, 2, id, "2025-02-01")
		ts.rate(t, 2, id, 2, "2025-02-01")

		assert.Equal(t, []ratedTitle{{"Shared", 8}}, ratedTitles(ts.bestOfTheYear(t)))
	})

	t.Run("uses the latest rating of a movie rated again", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.watchedAndRated(t, "Rated again", 6, "2024-02-01")
		ts.rate(t, 1, id, 10, "2024-09-01")
		ts.watchedAndRated(t, "Nine", 9, "2024-01-01")

		assert.Equal(t, []ratedTitle{{"Rated again", 10}, {"Nine", 9}}, ratedTitles(ts.bestOfTheYear(t)))
	})

	t.Run("uses the latest rating even when it is lower", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.watchedAndRated(t, "Rated down", 9, "2024-02-01")
		ts.rate(t, 1, id, 4, "2024-09-01")
		ts.watchedAndRated(t, "Six", 6, "2024-01-01")

		assert.Equal(t, []ratedTitle{{"Six", 6}, {"Rated down", 4}}, ratedTitles(ts.bestOfTheYear(t)))
	})

	t.Run("uses the last added rating when two share a creation time", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.watchedAndRated(t, "Same time", 5, "2024-02-01")
		ts.rate(t, 1, id, 8, "2024-02-01")

		assert.Equal(t, []ratedTitle{{"Same time", 8}}, ratedTitles(ts.bestOfTheYear(t)))
	})

	t.Run("orders movies with the same rating and watch time by when they were added", func(t *testing.T) {
		ts := setupStats(t)

		for i := 1; i <= 11; i++ {
			ts.watchedAndRated(t, fmt.Sprintf("Tied %02d", i), 7, "2024-02-01")
		}

		titles := ratedTitles(ts.bestOfTheYear(t))

		assert.Len(t, titles, 10)
		assert.Equal(t, ratedTitle{"Tied 01", 7}, titles[0])
		assert.Equal(t, ratedTitle{"Tied 10", 7}, titles[9])
	})
}
