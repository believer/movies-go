package db

import (
	"believer/movies/db/dbtest"
	"believer/movies/types"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupWatched returns a fixture where user 1 watched "Rewatched" twice and
// only user 2 watched "Others only". Both movies share a genre, language,
// company, country, list, series, director, award nomination and win, and a
// rating of 8 by user 1, so every query that reads watched_movie sees both.
func setupWatched(t *testing.T) testStats {
	t.Helper()

	ts := setupStats(t)
	dbtest.Truncate(t, ts.db, "genre", "language", "production_company", "production_country", "series", "official_list")

	ts.exec(t, `INSERT INTO genre (id, name) VALUES (1, 'Drama')`)
	ts.exec(t, `INSERT INTO language (id, name, english_name) VALUES (1, 'Svenska', 'Swedish')`)
	ts.exec(t, `INSERT INTO production_company (id, tmdb_id, name) VALUES (1, 1, 'Company')`)
	ts.exec(t, `INSERT INTO production_country (id, name) VALUES ('SE', 'Sweden')`)
	ts.exec(t, `INSERT INTO series (id, name) VALUES (1, 'Series')`)
	ts.exec(t, `INSERT INTO official_list (id, slug, name, source) VALUES (1, 'list', 'List', 'test')`)

	rewatched := ts.movie(t, "Rewatched")
	othersOnly := ts.movie(t, "Others only")

	for i, id := range []int{rewatched, othersOnly} {
		ts.exec(t, `UPDATE movie SET release_date = '2001-01-01', wilhelm = TRUE WHERE id = $1`, id)
		ts.exec(t, `INSERT INTO movie_genre (movie_id, genre_id) VALUES ($1, 1)`, id)
		ts.exec(t, `INSERT INTO movie_language (movie_id, language_id) VALUES ($1, 1)`, id)
		ts.exec(t, `INSERT INTO movie_company (movie_id, company_id) VALUES ($1, 1)`, id)
		ts.exec(t, `INSERT INTO movie_country (movie_id, country_id) VALUES ($1, 'SE')`, id)
		ts.exec(t, `INSERT INTO movie_series (movie_id, series_id, number_in_series) VALUES ($1, 1, $2)`, id, i+1)
		ts.exec(t, `INSERT INTO official_list_movie (list_id, movie_id, rank) VALUES (1, $1, $2)`, id, i+1)
		ts.exec(t, `INSERT INTO award (id, imdb_id, name, year, winner) SELECT $1, imdb_id, 'Best Picture', '2001', TRUE FROM movie WHERE id = $2`, i+1, id)
		ts.rate(t, 1, id, 8, "2024-01-01")
	}

	ts.director(t, "Director", rewatched, othersOnly)

	ts.watch(t, 1, rewatched, "2024-01-01")
	ts.watch(t, 1, rewatched, "2024-06-01")
	ts.watch(t, 2, othersOnly, "2024-01-01")

	return ts
}

type seenTitle struct {
	Title string
	Seen  bool
}

func seenTitles(movies types.Movies) []seenTitle {
	titles := []seenTitle{}

	for _, m := range movies {
		titles = append(titles, seenTitle{m.Title, m.Seen})
	}

	return titles
}

func TestWatchedMovie(t *testing.T) {
	// Each movie appears once, and only user 1's Watches mark it seen
	both := []seenTitle{{"Rewatched", true}, {"Others only", false}}

	movieLists := []struct {
		name  string
		query func(ts testStats) (types.Movies, error)
	}{
		{"GetByNominations", func(ts testStats) (types.Movies, error) {
			return NewAwardsRepository(ts.db).GetByNominations("1", 1, "academy-award")
		}},
		{"GetByWins", func(ts testStats) (types.Movies, error) {
			return NewAwardsRepository(ts.db).GetByWins("1", 1, "academy-award")
		}},
		{"GetListMovies", func(ts testStats) (types.Movies, error) {
			return NewListRepository(ts.db).GetListMovies("1", "1")
		}},
		{"GetGenreMovies", func(ts testStats) (types.Movies, error) {
			return NewGenreRepository(ts.db).GetGenreMovies("1", "1", 0)
		}},
		{"GetLanguageMovies", func(ts testStats) (types.Movies, error) {
			return NewLanguageRepository(ts.db).GetLanguageMovies("1", "1", 0)
		}},
		{"GetProductionCompanyMovies", func(ts testStats) (types.Movies, error) {
			return NewProductionCompanyRepository(ts.db).GetProductionCompanyMovies("1", "1", 0)
		}},
		{"GetProductionCountryMovies", func(ts testStats) (types.Movies, error) {
			return NewProductionCountryRepository(ts.db).GetProductionCountryMovies("SE", "1", 0)
		}},
		{"GetMoviesByRating", func(ts testStats) (types.Movies, error) {
			return NewRatingsRepository(ts.db).GetMoviesByRating("1", 8, 0)
		}},
		{"GetMoviesByYear", func(ts testStats) (types.Movies, error) {
			return NewYearsRepository(ts.db).GetMoviesByYear("1", "2001", 0)
		}},
	}

	for _, tt := range movieLists {
		t.Run(tt.name+" marks the user's watched movies seen, once each", func(t *testing.T) {
			ts := setupWatched(t)

			movies, err := tt.query(ts)

			if err != nil {
				t.Fatal(err)
			}

			assert.ElementsMatch(t, both, seenTitles(movies))
		})
	}

	t.Run("GetSeriesMovies marks the user's watched movies seen, once each", func(t *testing.T) {
		ts := setupWatched(t)

		series, err := NewSeriesRepository(ts.db).GetSeriesMovies("1", "1")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, series, 1)
		assert.Equal(t, both, seenTitles(types.Movies(series[0].Movies)))
	})

	t.Run("GetPersonByID marks the user's watched movies seen, once each", func(t *testing.T) {
		ts := setupWatched(t)

		var personID int

		if err := ts.db.Get(&personID, `SELECT id FROM person WHERE name = 'Director'`); err != nil {
			t.Fatal(err)
		}

		person, err := NewPersonRepository(ts.db).GetPersonByID(strconv.Itoa(personID), "1")

		if err != nil {
			t.Fatal(err)
		}

		titles := []seenTitle{}

		for _, m := range person.Director {
			titles = append(titles, seenTitle{m.Title, m.Seen})
		}

		assert.ElementsMatch(t, both, titles)
	})

	t.Run("GetLists counts a rewatched movie once and ignores other users' Watches", func(t *testing.T) {
		ts := setupWatched(t)

		lists, err := NewListRepository(ts.db).GetLists("1")

		if err != nil {
			t.Fatal(err)
		}

		assert.Len(t, lists, 1)
		assert.Equal(t, 2, lists[0].TotalMovies)
		assert.Equal(t, 1, lists[0].SeenMovies)
		assert.Equal(t, 50.0, lists[0].PercentSeen)
	})

	t.Run("GetWilhelmMovies lists the user's watched movies once each", func(t *testing.T) {
		ts := setupWatched(t)

		movies, err := NewStatsRepository(ts.db).GetWilhelmMovies("1", 0)

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, []seenTitle{{"Rewatched", true}}, seenTitles(movies))
	})

	t.Run("GetWilhelmScreamCount counts a rewatched movie once and ignores other users' Watches", func(t *testing.T) {
		ts := setupWatched(t)

		count, err := NewStatsRepository(ts.db).GetWilhelmScreamCount("1")

		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, []int{1}, count)
	})
}
