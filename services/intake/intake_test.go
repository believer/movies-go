package intake

import (
	"believer/movies/db/dbtest"
	"believer/movies/services/tmdb/tmdbtest"
	"believer/movies/types"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr(s string) *string { return &s }

var inception = types.MovieDetailsResponse{
	TmdbId:              27205,
	ImdbId:              "tt1375666",
	Title:               "Inception",
	Runtime:             148,
	ReleaseDate:         "2010-07-15",
	Genres:              []types.TmdbGenre{{Name: "Science Fiction"}, {Name: "Action"}},
	SpokenLanguages:     []types.TmdbLanguage{{ISO639: "en", EnglishName: "English", Name: "English"}},
	ProductionCountries: []types.TmdbCountry{{ID: "US", Name: "United States of America"}},
	ProductionCompanies: []types.TmdbCompany{{ID: 923, Name: "Legendary Pictures", OriginCountry: "US"}},
}

var inceptionCredits = types.MovieCreditsResponse{
	Cast: []types.CastResult{{ID: 6193, Name: "Leonardo DiCaprio", Character: ptr("Cobb")}},
	Crew: []types.CrewResult{
		{Id: 525, Name: "Christopher Nolan", Department: "Directing", Job: ptr("Director")},
		{Id: 559, Name: "Wally Pfister", Department: "Camera", Job: ptr("Director of Photography")},
	},
}

func setup(t *testing.T) (*Intake, *tmdbtest.Fake, *sqlx.DB) {
	t.Helper()

	db := dbtest.Open(t)
	dbtest.Truncate(t, db, "movie", "person", "genre", "language", "production_company", "production_country", "series")

	_, err := db.Exec(`INSERT INTO production_country (id, name) VALUES ('US', 'United States of America')`)
	require.NoError(t, err)

	// Awards are read from the CSV files at the repository root
	t.Chdir("../..")

	fake := &tmdbtest.Fake{}
	fake.Add(inception, inceptionCredits, "tt1375666", "27205")

	return New(db, fake), fake, db
}

func count(t *testing.T, db *sqlx.DB, query string, args ...any) int {
	t.Helper()

	var n int
	require.NoError(t, db.Get(&n, query, args...))

	return n
}

func TestEnsureMovie(t *testing.T) {
	t.Run("adds a missing movie with its metadata and awards", func(t *testing.T) {
		in, _, db := setup(t)

		id, err := in.EnsureMovie(ByExternalID("tt1375666"), MovieOptions{WilhelmScream: true})
		require.NoError(t, err)

		var movie struct {
			Title   string `db:"title"`
			TmdbID  int    `db:"tmdb_id"`
			Wilhelm bool   `db:"wilhelm"`
		}
		require.NoError(t, db.Get(&movie, `SELECT title, tmdb_id, wilhelm FROM movie WHERE id = $1`, id))
		assert.Equal(t, "Inception", movie.Title)
		assert.Equal(t, 27205, movie.TmdbID)
		assert.True(t, movie.Wilhelm)

		assert.Equal(t, 2, count(t, db, `SELECT count(*) FROM movie_genre WHERE movie_id = $1`, id))
		assert.Equal(t, 1, count(t, db, `SELECT count(*) FROM movie_language WHERE movie_id = $1`, id))
		assert.Equal(t, 1, count(t, db, `SELECT count(*) FROM movie_country WHERE movie_id = $1`, id))
		assert.Equal(t, 1, count(t, db, `SELECT count(*) FROM movie_company WHERE movie_id = $1`, id))
		assert.Equal(t, 3, count(t, db, `SELECT count(*) FROM movie_person WHERE movie_id = $1`, id))

		// The cinematography award is matched to the crew member
		assert.Equal(t, 1, count(t, db, `
SELECT count(*) FROM award a
    INNER JOIN person p ON p.id = a.person_id
WHERE a.imdb_id = 'tt1375666' AND a.name = 'Cinematography' AND p.name = 'Wally Pfister' AND a.winner`))
	})

	t.Run("finds an existing movie by IMDb or TMDB ID without fetching", func(t *testing.T) {
		in, fake, _ := setup(t)

		id, err := in.EnsureMovie(ByExternalID("tt1375666"), MovieOptions{})
		require.NoError(t, err)

		byImdb, err := in.EnsureMovie(ByExternalID("tt1375666"), MovieOptions{})
		require.NoError(t, err)

		byTmdb, err := in.EnsureMovie(ByExternalID("27205"), MovieOptions{})
		require.NoError(t, err)

		byID, err := in.EnsureMovie(ByID(id), MovieOptions{})
		require.NoError(t, err)

		assert.Equal(t, []int{id, id, id}, []int{byImdb, byTmdb, byID})
		assert.Equal(t, []string{"tt1375666"}, fake.Calls())
	})

	t.Run("returns ErrMovieNotFound for unknown movies", func(t *testing.T) {
		in, _, db := setup(t)

		_, err := in.EnsureMovie(ByExternalID("tt0000000"), MovieOptions{})
		assert.ErrorIs(t, err, ErrMovieNotFound)

		_, err = in.EnsureMovie(ByID(999), MovieOptions{})
		assert.ErrorIs(t, err, ErrMovieNotFound)

		assert.Equal(t, 0, count(t, db, `SELECT count(*) FROM movie`))
	})

	t.Run("adds the movie to a new or existing series", func(t *testing.T) {
		in, _, db := setup(t)

		id, err := in.EnsureMovie(ByExternalID("tt1375666"), MovieOptions{
			Series: &SeriesEntry{Name: "Dreams", Number: 1},
		})
		require.NoError(t, err)

		var seriesID int
		require.NoError(t, db.Get(&seriesID, `SELECT id FROM series WHERE name = 'Dreams'`))

		_, err = in.EnsureMovie(ByID(id), MovieOptions{
			Series: &SeriesEntry{ID: seriesID, Number: 2.5},
		})
		require.NoError(t, err)

		var numbers []float64
		require.NoError(t, db.Select(&numbers, `SELECT number_in_series FROM movie_series WHERE movie_id = $1 ORDER BY 1`, id))
		assert.Equal(t, []float64{1, 2.5}, numbers)
	})

	t.Run("returns ErrAlreadyInSeries when the movie has that number", func(t *testing.T) {
		in, _, db := setup(t)

		entry := MovieOptions{Series: &SeriesEntry{Name: "Nolan", Number: 1}}

		_, err := in.EnsureMovie(ByExternalID("tt1375666"), entry)
		require.NoError(t, err)

		_, err = in.EnsureMovie(ByExternalID("tt1375666"), entry)
		assert.ErrorIs(t, err, ErrAlreadyInSeries)

		assert.Equal(t, 1, count(t, db, `SELECT count(*) FROM movie_series`))
	})
}

func TestRefreshMovie(t *testing.T) {
	t.Run("overwrites details from TMDB", func(t *testing.T) {
		in, fake, db := setup(t)

		id, err := in.EnsureMovie(ByExternalID("tt1375666"), MovieOptions{})
		require.NoError(t, err)

		updated := inception
		updated.Title = "Inception (2010)"
		updated.Genres = append(updated.Genres, types.TmdbGenre{Name: "Thriller"})
		fake.Add(updated, inceptionCredits, "tt1375666")

		require.NoError(t, in.RefreshMovie(id))

		var title string
		require.NoError(t, db.Get(&title, `SELECT title FROM movie WHERE id = $1`, id))
		assert.Equal(t, "Inception (2010)", title)
		assert.Equal(t, 3, count(t, db, `SELECT count(*) FROM movie_genre WHERE movie_id = $1`, id))

		// Awards are upserted, not duplicated
		assert.Equal(t, 1, count(t, db, `SELECT count(*) FROM award WHERE imdb_id = 'tt1375666' AND name = 'Cinematography'`))
	})

	t.Run("returns ErrMovieNotFound for unknown movie", func(t *testing.T) {
		in, _, _ := setup(t)

		assert.ErrorIs(t, in.RefreshMovie(999), ErrMovieNotFound)
	})
}
