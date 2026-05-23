package handlers

import (
	"believer/movies/components/graph"
	"believer/movies/mocks"
	"believer/movies/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupStatsApp(h *StatsHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", false)
		return c.Next()
	})

	app.Get("/stats", h.GetStats)
	app.Get("/stats/best-of-the-year", h.GetBestOfTheYear)
	app.Get("/stats/wilhelm-scream", h.GetWilhelmScream)
	app.Get("/stats/seen-with", h.GetSeenWith)

	return app
}

func TestGetStats(t *testing.T) {
	repo := mocks.NewMockStatsQuerier(t)

	repo.On("GetReviewsCount", "user-123").Return(5, nil)
	repo.On("GetStatsData", "user-123").Return(types.Stats{
		UniqueMovies:      10,
		SeenWithRewatches: 12,
		TotalRuntime:      1200,
	}, nil)
	repo.On("GetMostWatchedByJob", "cast", "user-123", "All").Return([]types.ListItem{}, nil)
	repo.On("GetMostWatchedMovies", "user-123").Return([]types.ListItem{}, nil)
	repo.On("GetMoviesByYear", "user-123", "All").Return([]graph.GraphData{}, nil)
	repo.On("GetRatings", "user-123").Return([]graph.GraphData{}, nil)
	repo.On("GetWatchedThisYearByMonth", "user-123", mock.AnythingOfType("string")).Return([]graph.GraphData{}, nil)
	repo.On("GetShortestAndLongestMovie", "user-123").Return(types.Movies{}, nil)
	repo.On("GetTotalWatchedByJobAndYear", "user-123", "cast", "All").Return([]types.ListItem{}, nil)
	repo.On("GetWatchedByYear", "user-123").Return([]graph.GraphData{}, nil)
	repo.On("GetWatchedByWeekday", "user-123", "All").Return([]graph.GraphData{}, nil)
	repo.On("GetWilhelmScreamCount", "user-123").Return([]int{2}, nil)
	repo.On("GetRatingsThisYear", "user-123", mock.AnythingOfType("string")).Return([]graph.GraphData{}, nil)
	repo.On("GetMostAwardNominations", "user-123").Return(types.AwardPersonStat{}, nil)
	repo.On("GetMostAwardWins", "user-123").Return(types.AwardPersonStat{}, nil)
	repo.On("GetTopAwardedMovies", "user-123").Return([]types.AwardMovieStat{}, nil)

	app := setupStatsApp(NewStatsHandler(repo))

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetBestOfTheYear(t *testing.T) {
	repo := mocks.NewMockStatsQuerier(t)
	mockMovies := []types.ListItem{
		{ID: "1", Name: "The Godfather", Count: 10},
	}
	repo.On("GetBestOfTheYear", "user-123", mock.AnythingOfType("string")).Return(mockMovies, nil)

	app := setupStatsApp(NewStatsHandler(repo))

	req := httptest.NewRequest(http.MethodGet, "/stats/best-of-the-year", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetWilhelmScream(t *testing.T) {
	repo := mocks.NewMockStatsQuerier(t)
	repo.On("GetWilhelmMovies", "user-123", 0).Return(types.Movies{}, nil)

	app := setupStatsApp(NewStatsHandler(repo))

	req := httptest.NewRequest(http.MethodGet, "/stats/wilhelm-scream", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetSeenWith(t *testing.T) {
	repo := mocks.NewMockStatsQuerier(t)
	repo.On("GetSeenWith").Return([]types.ListItem{}, nil)

	app := setupStatsApp(NewStatsHandler(repo))

	req := httptest.NewRequest(http.MethodGet, "/stats/seen-with", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
