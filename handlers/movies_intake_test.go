package handlers

import (
	"believer/movies/mocks"
	"believer/movies/services/intake"
	"believer/movies/types"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var shawshank = types.MovieDetailsResponse{ImdbId: "tt0111161", TmdbId: 278, Title: "The Shawshank Redemption"}

// setupIntakeMovieApp signs in as user 1, a real user in the test database.
func setupIntakeMovieApp(h *MovieHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "1")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Post("/movie/new", h.PostMovieNew)
	app.Post("/movie/:id/seen", h.CreateSeenMovie)
	app.Post("/hooks/playback", h.PlaybackProgress)

	return app
}

func postForm(t *testing.T, app *fiber.App, path string, form url.Values) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	require.NoError(t, err)

	return resp
}

func TestPostMovieNew(t *testing.T) {
	t.Run("logs a watch of a new movie", func(t *testing.T) {
		movies := setupIntake(t)
		movies.seedUser(t, 1)
		movies.tmdb.Add(shawshank, types.MovieCreditsResponse{}, "tt0111161")

		app := setupIntakeMovieApp(NewMovieHandler(mocks.NewMockMovieQuerier(t), movies.intake))

		resp := postForm(t, app, "/movie/new", url.Values{
			"imdb_id":    {"https://www.imdb.com/title/tt0111161/"},
			"watched_at": {"2026-10-20"},
			"rating":     {"10"},
		})

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "/movie/1", resp.Header.Get("HX-Redirect"))
		assert.Equal(t, 1, movies.count(t, `SELECT count(*) FROM seen WHERE movie_id = 1 AND user_id = 1 AND date::date = '2026-10-20'`))
		assert.Equal(t, 1, movies.count(t, `SELECT count(*) FROM rating WHERE movie_id = 1 AND rating = 10`))
	})

	t.Run("adds to watchlist without a watch or rating", func(t *testing.T) {
		movies := setupIntake(t)
		movies.seedUser(t, 1)
		movies.tmdb.Add(shawshank, types.MovieCreditsResponse{}, "tt0111161")

		app := setupIntakeMovieApp(NewMovieHandler(mocks.NewMockMovieQuerier(t), movies.intake))
		form := url.Values{"imdb_id": {"tt0111161"}, "watchlist": {"true"}, "rating": {"8"}}

		resp := postForm(t, app, "/movie/new", form)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 1, movies.count(t, `SELECT count(*) FROM watchlist WHERE movie_id = 1`))
		assert.Equal(t, 0, movies.count(t, `SELECT count(*) FROM seen`))
		assert.Equal(t, 0, movies.count(t, `SELECT count(*) FROM rating`))

		resp = postForm(t, app, "/movie/new", form)
		assert.Equal(t, "#error", resp.Header.Get("HX-Retarget"))
	})

	t.Run("dry run only adds the movie", func(t *testing.T) {
		movies := setupIntake(t)
		movies.seedUser(t, 1)
		movies.tmdb.Add(shawshank, types.MovieCreditsResponse{}, "tt0111161")

		app := setupIntakeMovieApp(NewMovieHandler(mocks.NewMockMovieQuerier(t), movies.intake))

		resp := postForm(t, app, "/movie/new", url.Values{"imdb_id": {"tt0111161"}, "dry_run": {"true"}, "review": {"Hope"}})

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 1, movies.count(t, `SELECT count(*) FROM movie`))
		assert.Equal(t, 0, movies.count(t, `SELECT count(*) FROM seen`))
		assert.Equal(t, 0, movies.count(t, `SELECT count(*) FROM review`))
	})
}

func TestCreateSeenMovie(t *testing.T) {
	t.Run("logs another watch and clears the watchlist", func(t *testing.T) {
		movies := setupIntake(t)
		movies.seedUser(t, 1)
		movies.tmdb.Add(shawshank, types.MovieCreditsResponse{}, "tt0111161")

		id, err := movies.intake.AddToWatchlist("1", intake.ByExternalID("tt0111161"), intake.MovieOptions{})
		require.NoError(t, err)

		app := setupIntakeMovieApp(NewMovieHandler(mocks.NewMockMovieQuerier(t), movies.intake))

		resp := postForm(t, app, "/movie/1/seen", url.Values{})

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 1, movies.count(t, `SELECT count(*) FROM seen WHERE movie_id = $1`, id))
		assert.Equal(t, 0, movies.count(t, `SELECT count(*) FROM watchlist`))
	})
}

func TestPlaybackProgress(t *testing.T) {
	t.Run("logs a watch when playback completes", func(t *testing.T) {
		movies := setupIntake(t)
		movies.seedUser(t, 1)
		movies.tmdb.Add(shawshank, types.MovieCreditsResponse{}, "tt0111161")

		app := setupIntakeMovieApp(NewMovieHandler(mocks.NewMockMovieQuerier(t), movies.intake))

		resp := postForm(t, app, "/hooks/playback", url.Values{"imdb_id": {"tt0111161"}, "completed": {"true"}})

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 1, movies.count(t, `SELECT count(*) FROM seen WHERE user_id = 1`))
	})

	t.Run("updates now playing without fetching an existing movie", func(t *testing.T) {
		movies := setupIntake(t)
		movies.seedUser(t, 1)
		movies.tmdb.Add(shawshank, types.MovieCreditsResponse{}, "tt0111161")

		id, err := movies.intake.EnsureMovie(intake.ByExternalID("tt0111161"), intake.MovieOptions{})
		require.NoError(t, err)

		repo := mocks.NewMockMovieQuerier(t)
		repo.On("UpdateNowPlaying", id, 90.5, "1").Return(nil)

		app := setupIntakeMovieApp(NewMovieHandler(repo, movies.intake))

		for range 3 {
			resp := postForm(t, app, "/hooks/playback", url.Values{"imdb_id": {"tt0111161"}, "position": {"1:30:30"}})
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		assert.Equal(t, []string{"tt0111161"}, movies.tmdb.Calls())
	})
}
