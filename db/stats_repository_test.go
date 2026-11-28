package db

import (
	"believer/movies/components/graph"
	"believer/movies/db/dbtest"
	"believer/movies/types"
	"database/sql"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testStats struct {
	db *sqlx.DB
}

// setupStats returns a stats repository fixture backed by an empty test
// database with users 1 and 2. Tests using it skip with -short.
func setupStats(t *testing.T) testStats {
	t.Helper()

	db := dbtest.Open(t)
	dbtest.Truncate(t, db, "movie", "user", "seen", "rating", "person")

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

// person adds a person and returns their ID. Each gets the next original_id,
// which is unique.
func (ts testStats) person(t *testing.T, name string, popularity float64) int {
	t.Helper()

	var id int

	if err := ts.db.Get(&id, `INSERT INTO person (name, original_id, popularity) SELECT $1, COALESCE(MAX(original_id), 0) + 1, $2 FROM person RETURNING id`, name, popularity); err != nil {
		t.Fatal(err)
	}

	return id
}

// credit gives a person a job on a movie.
func (ts testStats) credit(t *testing.T, movieID int, personID int, job string) {
	t.Helper()
	ts.exec(t, `INSERT INTO movie_person (movie_id, person_id, job) VALUES ($1, $2, $3)`, movieID, personID, job)
}

// director adds a person who directed the given movies and returns their ID.
func (ts testStats) director(t *testing.T, name string, movieIDs ...int) int {
	t.Helper()

	id := ts.person(t, name, 0)

	for _, movieID := range movieIDs {
		ts.credit(t, movieID, id, "director")
	}

	return id
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

// ratingCounts maps each Rating with movies to how many it has.
func ratingCounts(t *testing.T, data []graph.GraphData, err error) map[string]int {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}

	counts := map[string]int{}

	for _, d := range data {
		if d.Value > 0 {
			counts[d.Label] = d.Value
		}
	}

	return counts
}

func TestGetRatings(t *testing.T) {
	t.Run("counts a movie rated again once, under its latest rating", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.watchedAndRated(t, "Rated down", 10, "2024-02-01")
		ts.rate(t, 1, id, 9, "2024-09-01")
		ts.watchedAndRated(t, "Nine", 9, "2024-03-01")
		ts.rate(t, 2, id, 3, "2025-01-01")

		data, err := NewStatsRepository(ts.db).GetRatings("1")

		assert.Equal(t, map[string]int{"9": 2}, ratingCounts(t, data, err))
	})
}

func TestGetRatingsThisYear(t *testing.T) {
	t.Run("counts a movie rated again once, under its latest rating", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.watchedAndRated(t, "Rated down", 10, "2024-02-01")
		ts.rate(t, 1, id, 9, "2024-09-01")
		ts.watchedAndRated(t, "Eight", 8, "2024-03-01")
		ts.rate(t, 2, id, 3, "2025-01-01")

		data, err := NewStatsRepository(ts.db).GetRatingsThisYear("1", "2024-01-01")

		assert.Equal(t, map[string]int{"9": 1, "8": 1}, ratingCounts(t, data, err))
	})

	t.Run("counts a movie watched twice in the year once", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.watchedAndRated(t, "Twice", 7, "2024-02-01")
		ts.watch(t, 1, id, "2024-08-01")

		data, err := NewStatsRepository(ts.db).GetRatingsThisYear("1", "2024-01-01")

		assert.Equal(t, map[string]int{"7": 1}, ratingCounts(t, data, err))
	})

	t.Run("leaves out movies not watched in the year", func(t *testing.T) {
		ts := setupStats(t)

		ts.watchedAndRated(t, "Last year", 7, "2023-02-01")

		data, err := NewStatsRepository(ts.db).GetRatingsThisYear("1", "2024-01-01")

		assert.Empty(t, ratingCounts(t, data, err))
	})
}

func TestGetHighestRankedPersonByJob(t *testing.T) {
	t.Run("counts a movie rated again once, with its latest rating", func(t *testing.T) {
		ts := setupStats(t)

		rated := ts.watchedAndRated(t, "Rated down", 10, "2024-02-01")
		ts.rate(t, 1, rated, 6, "2024-09-01")
		eight := ts.watchedAndRated(t, "Eight", 8, "2024-03-01")
		ts.rate(t, 2, eight, 1, "2025-01-01")
		ts.director(t, "Director", rated, eight)

		persons, err := NewStatsRepository(ts.db).GetHighestRankedPersonByJob("1", "director")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, persons, 1)
		assert.Equal(t, "Director", persons[0].Name)
		assert.Equal(t, 2, persons[0].Appearances)
		assert.Equal(t, 14, persons[0].TotalRating)
	})

	t.Run("ranks by average weighted by appearances, top 10", func(t *testing.T) {
		ts := setupStats(t)

		// 8 × log(3) beats 10 × log(2), so ranking by plain average fails
		ts.director(t, "Steady",
			ts.watchedAndRated(t, "Eight 1", 8, "2024-01-01"),
			ts.watchedAndRated(t, "Eight 2", 8, "2024-01-02"),
			ts.watchedAndRated(t, "Eight 3", 8, "2024-01-03"))
		ts.director(t, "Brilliant",
			ts.watchedAndRated(t, "Ten 1", 10, "2024-02-01"),
			ts.watchedAndRated(t, "Ten 2", 10, "2024-02-02"))

		// Ten more, ranked below, so the cutoff leaves the last one out
		for i := range 10 {
			ts.director(t, fmt.Sprintf("Fair %d", i),
				ts.watchedAndRated(t, fmt.Sprintf("Fair %d a", i), 10-i, "2024-03-01"),
				ts.watchedAndRated(t, fmt.Sprintf("Fair %d b", i), 1, "2024-03-02"))
		}

		persons, err := NewStatsRepository(ts.db).GetHighestRankedPersonByJob("1", "director")

		require.NoError(t, err)
		require.Len(t, persons, 10)
		assert.Equal(t, "Steady", persons[0].Name)
		assert.Equal(t, 24, persons[0].TotalRating)
		assert.InDelta(t, 8*math.Log10(3), persons[0].WeightedAverageRating, 1e-9)
		assert.Equal(t, "Brilliant", persons[1].Name)
		assert.Equal(t, "Fair 7", persons[9].Name)
	})

	t.Run("breaks ties on the weighted average by ID", func(t *testing.T) {
		ts := setupStats(t)

		tied := make([]int, 15)

		for i := range tied {
			tied[i] = ts.director(t, "Tied")
		}

		// Credited in reverse ID order, so the order of the rows doesn't
		// pass for the tiebreak
		for i := len(tied) - 1; i >= 0; i-- {
			ts.credit(t, ts.watchedAndRated(t, fmt.Sprintf("Tied %d a", i), 7, "2024-01-01"), tied[i], "director")
			ts.credit(t, ts.watchedAndRated(t, fmt.Sprintf("Tied %d b", i), 7, "2024-01-02"), tied[i], "director")
		}

		// With statistics the planner groups by hashing, as it does on
		// production data. Unanalyzed, it groups by sorting on person_id,
		// which returns ties in ID order without a tiebreak.
		ts.exec(t, `ANALYZE rating, movie_person`)

		persons, err := NewStatsRepository(ts.db).GetHighestRankedPersonByJob("1", "director")

		require.NoError(t, err)

		ids := make([]int, len(persons))

		for i, p := range persons {
			ids[i] = p.ID
		}

		// More ties than the top 10, so the cutoff picks the lowest IDs
		assert.Equal(t, tied[:10], ids)
	})
}

// imdbMovie adds a movie with an IMDb rating, stored as text like production
// stores it, and returns its ID.
func (ts testStats) imdbMovie(t *testing.T, title string, imdbRating any) int {
	t.Helper()

	var id int

	if err := ts.db.Get(&id, `INSERT INTO movie (title, imdb_id, imdb_rating) VALUES ($1, $1, $2) RETURNING id`, title, imdbRating); err != nil {
		t.Fatal(err)
	}

	return id
}

// ratedAgainstIMDb adds a movie with an IMDb rating that user 1 rated.
func (ts testStats) ratedAgainstIMDb(t *testing.T, title string, imdbRating string, rating int) int {
	t.Helper()

	id := ts.imdbMovie(t, title, imdbRating)
	ts.rate(t, 1, id, rating, "2024-01-01")

	return id
}

func hotTakeTitles(t *testing.T, takes []types.HotTake, err error) []string {
	t.Helper()
	require.NoError(t, err)

	titles := []string{}

	for _, take := range takes {
		titles = append(titles, take.Title)
	}

	return titles
}

func TestGetHotTakes(t *testing.T) {
	t.Run("returns the top 10 above IMDb, then the top 10 below", func(t *testing.T) {
		ts := setupStats(t)

		// Eleven each way, so the cutoff leaves the smallest one out
		for i := range 11 {
			ts.ratedAgainstIMDb(t, fmt.Sprintf("Above %02d", i), fmt.Sprintf("%.1f", 1+float64(i)/10), 10)
			ts.ratedAgainstIMDb(t, fmt.Sprintf("Below %02d", i), fmt.Sprintf("%.1f", 9.9-float64(i)/10), 1)
		}

		ts.ratedAgainstIMDb(t, "Agrees", "7.0", 7)

		takes, err := NewStatsRepository(ts.db).GetHotTakes("1")

		assert.Equal(t, []string{
			"Above 00", "Above 01", "Above 02", "Above 03", "Above 04",
			"Above 05", "Above 06", "Above 07", "Above 08", "Above 09",
			"Below 00", "Below 01", "Below 02", "Below 03", "Below 04",
			"Below 05", "Below 06", "Below 07", "Below 08", "Below 09",
		}, hotTakeTitles(t, takes, err))
	})

	t.Run("returns the ratings and their exact difference", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.ratedAgainstIMDb(t, "Vertical Limit", "5.9", 9)

		takes, err := NewStatsRepository(ts.db).GetHotTakes("1")

		require.NoError(t, err)
		assert.Equal(t, []types.HotTake{{
			ID:         id,
			Title:      "Vertical Limit",
			Rating:     9,
			IMDbRating: 5.9,
			Difference: 3.1,
		}}, takes)
	})

	t.Run("leaves out movies without an IMDb rating", func(t *testing.T) {
		ts := setupStats(t)

		for _, imdbRating := range []any{nil, "", "0"} {
			ts.rate(t, 1, ts.imdbMovie(t, fmt.Sprintf("No rating %v", imdbRating), imdbRating), 9, "2024-01-01")
		}

		ts.ratedAgainstIMDb(t, "Rated", "6.0", 9)

		takes, err := NewStatsRepository(ts.db).GetHotTakes("1")

		assert.Equal(t, []string{"Rated"}, hotTakeTitles(t, takes, err))
	})

	t.Run("uses the latest rating of a movie rated again", func(t *testing.T) {
		ts := setupStats(t)

		// The earlier rating is far above IMDb, the latest below it
		id := ts.ratedAgainstIMDb(t, "Rated down", "7.0", 10)
		ts.rate(t, 1, id, 5, "2024-09-01")

		takes, err := NewStatsRepository(ts.db).GetHotTakes("1")

		require.NoError(t, err)
		require.Len(t, takes, 1)
		assert.Equal(t, 5, takes[0].Rating)
		assert.InDelta(t, -2.0, takes[0].Difference, 1e-9)
	})

	t.Run("ignores other users' ratings", func(t *testing.T) {
		ts := setupStats(t)

		id := ts.imdbMovie(t, "Theirs", "5.0")
		ts.rate(t, 2, id, 10, "2024-01-01")

		takes, err := NewStatsRepository(ts.db).GetHotTakes("1")

		assert.Empty(t, hotTakeTitles(t, takes, err))
	})

	t.Run("breaks ties on the difference by ID", func(t *testing.T) {
		ts := setupStats(t)

		// More than twice the cutoff, so Postgres sorts with a bounded heap,
		// which doesn't keep the rows' order
		ids := make([]int, 25)

		for i := range ids {
			ids[i] = ts.imdbMovie(t, fmt.Sprintf("Tied %02d", i), "5.0")
		}

		// Rated in reverse ID order, so the order of the rows doesn't pass
		// for the tiebreak
		for i := len(ids) - 1; i >= 0; i-- {
			ts.rate(t, 1, ids[i], 8, "2024-01-01")
		}

		ts.exec(t, `ANALYZE rating, movie`)

		takes, err := NewStatsRepository(ts.db).GetHotTakes("1")

		assert.Equal(t, []string{
			"Tied 00", "Tied 01", "Tied 02", "Tied 03", "Tied 04",
			"Tied 05", "Tied 06", "Tied 07", "Tied 08", "Tied 09",
		}, hotTakeTitles(t, takes, err))
	})
}

func TestGetIMDbComparison(t *testing.T) {
	t.Run("averages the latest ratings minus IMDb's", func(t *testing.T) {
		ts := setupStats(t)

		ts.ratedAgainstIMDb(t, "Above", "6.5", 9)
		ts.ratedAgainstIMDb(t, "Below", "8.0", 7)
		id := ts.ratedAgainstIMDb(t, "Rated down", "7.0", 10)
		ts.rate(t, 1, id, 6, "2024-09-01")
		ts.rate(t, 1, ts.imdbMovie(t, "No rating", "0"), 1, "2024-01-01")
		ts.rate(t, 2, ts.imdbMovie(t, "Theirs", "5.0"), 10, "2024-01-01")

		comparison, err := NewStatsRepository(ts.db).GetIMDbComparison("1")

		require.NoError(t, err)
		assert.Equal(t, 3, comparison.Movies)
		// (2.5 - 1.0 - 1.0) / 3
		assert.InDelta(t, 0.5/3, comparison.AverageDifference, 1e-9)
	})

	t.Run("is zero without rated movies", func(t *testing.T) {
		ts := setupStats(t)

		comparison, err := NewStatsRepository(ts.db).GetIMDbComparison("1")

		require.NoError(t, err)
		assert.Equal(t, types.IMDbComparison{}, comparison)
	})
}

func (ts testStats) streaks(t *testing.T) types.Streaks {
	t.Helper()

	streaks, err := NewStatsRepository(ts.db).GetStreaks("1")
	require.NoError(t, err)

	return streaks
}

func day(t *testing.T, nt sql.NullTime) string {
	t.Helper()

	if !nt.Valid {
		return ""
	}

	return nt.Time.Format("2006-01-02")
}

func TestGetStreaks(t *testing.T) {
	t.Run("finds the most consecutive days with a Watch", func(t *testing.T) {
		ts := setupStats(t)
		id := ts.movie(t, "Movie")

		for _, date := range []string{"2024-01-01", "2024-01-02", "2024-01-03"} {
			ts.watch(t, 1, id, date)
		}

		// Four days over a year end, with two Watches on one of them, so
		// counting Watches rather than days gives the wrong streak
		for _, date := range []string{"2024-12-30", "2024-12-31", "2024-12-31 22:00", "2025-01-01", "2025-01-02"} {
			ts.watch(t, 1, id, date)
		}

		ts.watch(t, 1, id, "2025-03-01")

		s := ts.streaks(t)

		assert.Equal(t, 4, s.StreakDays)
		assert.Equal(t, "2024-12-30", day(t, s.StreakStart))
		assert.Equal(t, "2025-01-02", day(t, s.StreakEnd))
	})

	t.Run("finds the most days without a Watch between two", func(t *testing.T) {
		ts := setupStats(t)
		id := ts.movie(t, "Movie")

		for _, date := range []string{"2024-01-01", "2024-01-11", "2024-03-01 23:00", "2024-03-05"} {
			ts.watch(t, 1, id, date)
		}

		s := ts.streaks(t)

		// January 12 to February 29
		assert.Equal(t, 49, s.BreakDays)
		assert.Equal(t, "2024-01-11", day(t, s.BreakStart))
		assert.Equal(t, "2024-03-01", day(t, s.BreakEnd))
	})

	t.Run("finds the day with the most Watches", func(t *testing.T) {
		ts := setupStats(t)
		id := ts.movie(t, "Movie")

		for _, date := range []string{"2024-01-01", "2024-01-01 12:00", "2024-02-01 10:00", "2024-02-01 14:00", "2024-02-01 23:59"} {
			ts.watch(t, 1, id, date)
		}

		s := ts.streaks(t)

		assert.Equal(t, 3, s.BusiestDayWatches)
		assert.Equal(t, "2024-02-01", day(t, s.BusiestDay))
	})

	t.Run("breaks ties with the earliest record", func(t *testing.T) {
		ts := setupStats(t)
		id := ts.movie(t, "Movie")

		// Added latest first, so the order of the rows doesn't pass for the
		// tiebreak. Two-day streaks with two Watches on their first day, ten
		// days apart.
		for _, start := range []string{"2024-03-01", "2024-02-10", "2024-01-21", "2024-01-01"} {
			date, err := time.Parse("2006-01-02", start)
			require.NoError(t, err)

			ts.watch(t, 1, id, date.Format(time.DateOnly))
			ts.watch(t, 1, id, date.Format(time.DateOnly)+" 12:00")
			ts.watch(t, 1, id, date.AddDate(0, 0, 1).Format(time.DateOnly))
		}

		ts.exec(t, `ANALYZE seen`)

		s := ts.streaks(t)

		assert.Equal(t, "2024-01-01", day(t, s.StreakStart))
		assert.Equal(t, "2024-01-02", day(t, s.BreakStart))
		assert.Equal(t, "2024-01-01", day(t, s.BusiestDay))
	})

	t.Run("leaves out the 2011 bulk import", func(t *testing.T) {
		ts := setupStats(t)
		id := ts.movie(t, "Movie")

		for _, date := range []string{"2011-12-30", "2011-12-30 12:00", "2011-12-31", "2012-01-01"} {
			ts.watch(t, 1, id, date)
		}

		ts.watch(t, 1, id, "2012-02-01")

		s := ts.streaks(t)

		assert.Equal(t, 1, s.StreakDays)
		assert.Equal(t, "2012-01-01", day(t, s.StreakStart))
		assert.Equal(t, 1, s.BusiestDayWatches)
		assert.Equal(t, "2012-01-01", day(t, s.BreakStart))
	})

	t.Run("ignores other users' Watches", func(t *testing.T) {
		ts := setupStats(t)
		id := ts.movie(t, "Movie")

		ts.watch(t, 1, id, "2024-01-01")
		ts.watch(t, 1, id, "2024-01-05")

		for _, date := range []string{"2024-01-02", "2024-01-02 12:00", "2024-01-03"} {
			ts.watch(t, 2, id, date)
		}

		s := ts.streaks(t)

		assert.Equal(t, 1, s.StreakDays)
		assert.Equal(t, 3, s.BreakDays)
		assert.Equal(t, 1, s.BusiestDayWatches)
	})

	t.Run("has no break with a single day of Watches", func(t *testing.T) {
		ts := setupStats(t)
		ts.watch(t, 1, ts.movie(t, "Movie"), "2024-01-01")

		s := ts.streaks(t)

		assert.Equal(t, 1, s.StreakDays)
		assert.Equal(t, 0, s.BreakDays)
		assert.False(t, s.BreakStart.Valid)
	})

	t.Run("has no records without Watches", func(t *testing.T) {
		ts := setupStats(t)

		assert.Equal(t, types.Streaks{}, ts.streaks(t))
	})
}

// cast credits people as the movie's cast in billing order.
func (ts testStats) cast(t *testing.T, movieID int, personIDs ...int) {
	t.Helper()

	for _, personID := range personIDs {
		ts.credit(t, movieID, personID, "cast")
	}
}

func pairNames(t *testing.T, pairs []types.PersonPair, err error) []string {
	t.Helper()
	require.NoError(t, err)

	names := []string{}

	for _, p := range pairs {
		names = append(names, fmt.Sprintf("%s & %s (%d)", p.ActorName, p.DirectorName, p.Movies))
	}

	return names
}

func TestGetActorDirectorPairs(t *testing.T) {
	t.Run("counts the movies each pair made, most first", func(t *testing.T) {
		ts := setupStats(t)

		bale := ts.person(t, "Christian Bale", 0)
		caine := ts.person(t, "Michael Caine", 0)
		movies := []int{
			ts.watchedAndRated(t, "Batman Begins", 8, "2024-01-01"),
			ts.watchedAndRated(t, "The Dark Knight", 9, "2024-01-02"),
			ts.watchedAndRated(t, "The Prestige", 9, "2024-01-03"),
		}

		ts.director(t, "Christopher Nolan", movies...)
		ts.cast(t, movies[0], bale, caine)
		ts.cast(t, movies[1], bale, caine)
		ts.cast(t, movies[2], caine, bale)

		// A rewatch counts the movie once
		ts.watch(t, 1, movies[2], "2024-06-01")

		pairs, err := NewStatsRepository(ts.db).GetActorDirectorPairs("1")

		// Ordered by ID within a count, and Bale was added first
		assert.Equal(t, []string{
			"Christian Bale & Christopher Nolan (3)",
			"Michael Caine & Christopher Nolan (3)",
		}, pairNames(t, pairs, err))
	})

	t.Run("leaves out pairs with a single movie", func(t *testing.T) {
		ts := setupStats(t)

		once := ts.watchedAndRated(t, "Once", 8, "2024-01-01")
		ts.director(t, "Director", once)
		ts.cast(t, once, ts.person(t, "Actor", 0))

		pairs, err := NewStatsRepository(ts.db).GetActorDirectorPairs("1")

		assert.Empty(t, pairNames(t, pairs, err))
	})

	t.Run("counts the top 5 billed cast only", func(t *testing.T) {
		ts := setupStats(t)

		billed := make([]int, 6)

		for i := range billed {
			billed[i] = ts.person(t, fmt.Sprintf("Billed %d", i+1), 0)
		}

		first := ts.watchedAndRated(t, "First", 8, "2024-01-01")
		second := ts.watchedAndRated(t, "Second", 8, "2024-01-02")
		ts.director(t, "Director", first, second)
		ts.cast(t, first, billed...)
		// Billed sixth here too, but fifth would count
		ts.cast(t, second, billed...)

		pairs, err := NewStatsRepository(ts.db).GetActorDirectorPairs("1")

		assert.Equal(t, []string{
			"Billed 1 & Director (2)",
			"Billed 2 & Director (2)",
			"Billed 3 & Director (2)",
			"Billed 4 & Director (2)",
			"Billed 5 & Director (2)",
		}, pairNames(t, pairs, err))
	})

	t.Run("leaves out actors who directed themselves", func(t *testing.T) {
		ts := setupStats(t)

		first := ts.watchedAndRated(t, "First", 8, "2024-01-01")
		second := ts.watchedAndRated(t, "Second", 8, "2024-01-02")
		eastwood := ts.director(t, "Clint Eastwood", first, second)
		ts.cast(t, first, eastwood)
		ts.cast(t, second, eastwood)

		pairs, err := NewStatsRepository(ts.db).GetActorDirectorPairs("1")

		assert.Empty(t, pairNames(t, pairs, err))
	})

	t.Run("counts the user's watched movies only", func(t *testing.T) {
		ts := setupStats(t)

		actor := ts.person(t, "Actor", 0)
		mine := ts.watchedAndRated(t, "Mine", 8, "2024-01-01")
		theirs := ts.movie(t, "Theirs")
		unseen := ts.movie(t, "Unseen")
		ts.watch(t, 2, theirs, "2024-01-01")
		ts.director(t, "Director", mine, theirs, unseen)

		for _, id := range []int{mine, theirs, unseen} {
			ts.cast(t, id, actor)
		}

		pairs, err := NewStatsRepository(ts.db).GetActorDirectorPairs("1")

		assert.Empty(t, pairNames(t, pairs, err))
	})

	t.Run("returns the top 10, ties by actor and director ID", func(t *testing.T) {
		ts := setupStats(t)

		first := ts.watchedAndRated(t, "First", 8, "2024-01-01")
		second := ts.watchedAndRated(t, "Second", 8, "2024-01-02")
		directors := []int{
			ts.director(t, "Director A", first, second),
			ts.director(t, "Director B", first, second),
			ts.director(t, "Director C", first, second),
		}

		actors := make([]int, 3)

		for i := range actors {
			actors[i] = ts.person(t, fmt.Sprintf("Actor %d", i), 0)
		}

		ts.cast(t, first, actors...)
		ts.cast(t, second, actors...)

		// A third movie puts one pair on top, despite its higher IDs
		third := ts.watchedAndRated(t, "Third", 8, "2024-01-03")
		ts.credit(t, third, directors[1], "director")
		ts.cast(t, third, actors[2])

		// Two more actors make 15 pairs, so the cutoff leaves 5 out
		for i := range 2 {
			actor := ts.person(t, fmt.Sprintf("Later %d", i), 0)
			ts.cast(t, first, actor)
			ts.cast(t, second, actor)
		}

		ts.exec(t, `ANALYZE movie_person, seen`)

		pairs, err := NewStatsRepository(ts.db).GetActorDirectorPairs("1")

		assert.Equal(t, []string{
			"Actor 2 & Director B (3)",
			"Actor 0 & Director A (2)",
			"Actor 0 & Director B (2)",
			"Actor 0 & Director C (2)",
			"Actor 1 & Director A (2)",
			"Actor 1 & Director B (2)",
			"Actor 1 & Director C (2)",
			"Actor 2 & Director A (2)",
			"Actor 2 & Director C (2)",
			"Later 0 & Director A (2)",
		}, pairNames(t, pairs, err))
	})
}
