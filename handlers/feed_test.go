package handlers

import (
	"believer/movies/mocks"
	"believer/movies/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupFeedApp(feedRepo *mocks.MockFeedQuerier, nowPlayingRepo *mocks.MockNowPlayingQuerier) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", false)
		return c.Next()
	})

	h := NewFeedHandler(feedRepo, nowPlayingRepo)
	app.Get("/", h.GetFeed)

	return app
}

func TestGetFeed(t *testing.T) {
	t.Run("success home feed", func(t *testing.T) {
		feedRepo := mocks.NewMockFeedQuerier(t)
		nowPlayingRepo := mocks.NewMockNowPlayingQuerier(t)

		mockMovies := types.Movies{
			{ID: 1, Title: "Inception"},
		}
		mockNowPlaying := types.Movies{
			{ID: 2, Title: "Avatar 3"},
		}

		feedRepo.On("GetFeedMovies", "user-123", 0).Return(mockMovies, nil)
		nowPlayingRepo.On("GetNowPlaying", "user-123").Return(mockNowPlaying, nil)

		app := setupFeedApp(feedRepo, nowPlayingRepo)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("search movies", func(t *testing.T) {
		feedRepo := mocks.NewMockFeedQuerier(t)
		nowPlayingRepo := mocks.NewMockNowPlayingQuerier(t)

		mockMovies := types.Movies{
			{ID: 1, Title: "Inception"},
		}
		mockNowPlaying := types.Movies{}

		feedRepo.On("SearchMovies", "Inception").Return(mockMovies, nil)
		nowPlayingRepo.On("GetNowPlaying", "user-123").Return(mockNowPlaying, nil)

		app := setupFeedApp(feedRepo, nowPlayingRepo)

		req := httptest.NewRequest(http.MethodGet, "/?search=Inception", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
