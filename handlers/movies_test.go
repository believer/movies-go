package handlers

import (
	"believer/movies/components/movie"
	"believer/movies/mocks"
	"believer/movies/types"
	"believer/movies/views"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupMovieApp(h *MovieHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", c.Cookies("token") != "")
		return c.Next()
	})

	app.Get("/movie/imdb", h.GetByImdbId)
	app.Get("/movie/:id", h.GetMovieByID)
	app.Get("/movie/:id/seen/others", h.GetMovieOthersSeenByID)
	app.Delete("/movie/:id/now-playing", h.DeleteNowPlaying)

	return app
}

func TestGetMovieByID(t *testing.T) {
	t.Run("returns 200 with movie details", func(t *testing.T) {
		repo := mocks.NewMockMovieQuerier(t)
		
		mockMovie := types.Movie{
			ID:    456,
			Title: "Inception",
		}
		mockReview := types.Review{ID: 1, Content: "Good"}
		mockOthers := types.OthersStats{Seen: 2, AverageRating: 8.5}
		mockSeen := []movie.WatchedAt{{ID: 1}}
		mockCast := []views.CastDTO{{Job: "Director"}}

		repo.On("GetByID", "456", "user-123").Return(mockMovie, nil)
		repo.On("GetReviewByMovieID", "456", "user-123").Return(mockReview, nil)
		repo.On("IsWatchlisted", "456", "user-123").Return(false, nil)
		repo.On("RatingsByOthers", "456").Return(mockOthers, nil)
		repo.On("SeenByUser", "456", "user-123").Return(mockSeen, nil)
		repo.On("Cast", "456").Return(mockCast, false, nil)

		app := setupMovieApp(NewMovieHandler(repo))

		req := httptest.NewRequest(http.MethodGet, "/movie/inception-456", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestGetMovieOthersSeenByID(t *testing.T) {
	t.Run("returns 200", func(t *testing.T) {
		repo := mocks.NewMockMovieQuerier(t)
		mockOthers := types.OthersStats{Seen: 2, AverageRating: 8.5}
		repo.On("RatingsByOthers", "456").Return(mockOthers, nil)

		app := setupMovieApp(NewMovieHandler(repo))

		req := httptest.NewRequest(http.MethodGet, "/movie/inception-456/seen/others", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestGetByImdbId(t *testing.T) {
	t.Run("returns existing movie", func(t *testing.T) {
		repo := mocks.NewMockMovieQuerier(t)
		mockMovie := types.Movie{ID: 456, Title: "Inception"}
		repo.On("GetMovieByImdbID", "tt1375666").Return(mockMovie, nil)

		app := setupMovieApp(NewMovieHandler(repo))

		req := httptest.NewRequest(http.MethodGet, "/movie/imdb?imdbId=tt1375666", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestDeleteNowPlaying(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		repo := mocks.NewMockMovieQuerier(t)
		app := setupMovieApp(NewMovieHandler(repo))

		req := httptest.NewRequest(http.MethodDelete, "/movie/456/now-playing", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorized", func(t *testing.T) {
		repo := mocks.NewMockMovieQuerier(t)
		mockMovie := types.Movie{ID: 456, Title: "Inception", ImdbId: "tt1375666"}
		repo.On("GetMovieTitleAndImdbID", "456").Return(mockMovie, nil)
		repo.On("DeleteNowPlayingDirect", "user-123", "tt1375666").Return(nil)
		repo.On("SeenByUser", "456", "user-123").Return([]movie.WatchedAt{}, nil)
		repo.On("IsWatchlisted", "456", "user-123").Return(true, nil)

		app := setupMovieApp(NewMovieHandler(repo))

		req := httptest.NewRequest(http.MethodDelete, "/movie/456/now-playing", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
