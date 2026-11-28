package handlers

import (
	"believer/movies/components/graph"
	"believer/movies/db/mocks"
	"believer/movies/types"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	app.Get("/stats/hot-takes", h.GetHotTakes)
	app.Get("/stats/streaks", h.GetStreaks)
	app.Get("/stats/actor-director-pairs", h.GetActorDirectorPairs)

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

// getBody requests path and returns the response's body.
func getBody(t *testing.T, app *fiber.App, path string) string {
	t.Helper()

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return string(body)
}

func TestGetHotTakes(t *testing.T) {
	t.Run("splits the movies above and below IMDb", func(t *testing.T) {
		repo := mocks.NewMockStatsQuerier(t)
		repo.On("GetHotTakes", "user-123").Return([]types.HotTake{
			{ID: 1057, Title: "Vertical Limit", Rating: 9, IMDbRating: 5.9, Difference: 3.1},
			{ID: 1164, Title: "Amadeus", Rating: 2, IMDbRating: 8.3, Difference: -6.3},
		}, nil)
		repo.On("GetIMDbComparison", "user-123").Return(types.IMDbComparison{Movies: 2080, AverageDifference: -0.19}, nil)

		body := getBody(t, setupStatsApp(NewStatsHandler(repo)), "/stats/hot-takes")

		above := strings.Index(body, "Above IMDb")
		below := strings.Index(body, "Below IMDb")

		require.NotEqual(t, -1, above)
		require.NotEqual(t, -1, below)
		assert.Contains(t, body[above:below], `href="/movie/vertical-limit-1057"`)
		assert.Contains(t, body[above:below], "+3.1")
		assert.Contains(t, body[below:], `href="/movie/amadeus-1164"`)
		assert.Contains(t, body[below:], "−6.3")
		assert.Contains(t, body, "0.2 below IMDb on average")
		assert.Contains(t, body, "2,080 movies")
	})

	t.Run("shows an empty state without rated movies", func(t *testing.T) {
		repo := mocks.NewMockStatsQuerier(t)
		repo.On("GetHotTakes", "user-123").Return([]types.HotTake{}, nil)
		repo.On("GetIMDbComparison", "user-123").Return(types.IMDbComparison{}, nil)

		body := getBody(t, setupStatsApp(NewStatsHandler(repo)), "/stats/hot-takes")

		assert.Contains(t, body, "Hot takes")
		assert.Contains(t, body, "No rated movies with an IMDb rating")
		assert.NotContains(t, body, "Above IMDb")
	})
}

func date(t *testing.T, value string) sql.NullTime {
	t.Helper()

	d, err := time.Parse(time.DateOnly, value)
	require.NoError(t, err)

	return sql.NullTime{Time: d, Valid: true}
}

func TestGetStreaks(t *testing.T) {
	t.Run("shows the records with their dates", func(t *testing.T) {
		repo := mocks.NewMockStatsQuerier(t)
		repo.On("GetStreaks", "user-123").Return(types.Streaks{
			StreakDays:        10,
			StreakStart:       date(t, "2025-12-25"),
			StreakEnd:         date(t, "2026-01-03"),
			BreakDays:         155,
			BreakStart:        date(t, "2026-04-29"),
			BreakEnd:          date(t, "2026-10-02"),
			BusiestDayWatches: 8,
			BusiestDay:        date(t, "2012-07-20"),
		}, nil)

		body := getBody(t, setupStatsApp(NewStatsHandler(repo)), "/stats/streaks")

		assert.Contains(t, body, "10 days")
		assert.Contains(t, body, "December 25 2025 – January 3 2026")
		assert.Contains(t, body, "155 days")
		assert.Contains(t, body, "Between April 29 2026 and October 2 2026")
		assert.Contains(t, body, "8 movies")
		assert.Contains(t, body, "July 20 2012")
	})

	t.Run("leaves out the break with a single day of Watches", func(t *testing.T) {
		repo := mocks.NewMockStatsQuerier(t)
		repo.On("GetStreaks", "user-123").Return(types.Streaks{
			StreakDays:        1,
			StreakStart:       date(t, "2024-01-01"),
			StreakEnd:         date(t, "2024-01-01"),
			BusiestDayWatches: 1,
			BusiestDay:        date(t, "2024-01-01"),
		}, nil)

		body := getBody(t, setupStatsApp(NewStatsHandler(repo)), "/stats/streaks")

		assert.Contains(t, body, "Longest streak")
		assert.Contains(t, body, "1 day")
		assert.NotContains(t, body, "Longest break")
	})

	t.Run("shows an empty state without Watches", func(t *testing.T) {
		repo := mocks.NewMockStatsQuerier(t)
		repo.On("GetStreaks", "user-123").Return(types.Streaks{}, nil)

		body := getBody(t, setupStatsApp(NewStatsHandler(repo)), "/stats/streaks")

		assert.Contains(t, body, "No movies seen")
		assert.NotContains(t, body, "Longest streak")
	})
}

func TestGetActorDirectorPairs(t *testing.T) {
	t.Run("links both people of each pair", func(t *testing.T) {
		repo := mocks.NewMockStatsQuerier(t)
		repo.On("GetActorDirectorPairs", "user-123").Return([]types.PersonPair{
			{ActorID: 1, ActorName: "Christian Bale", DirectorID: 3, DirectorName: "Christopher Nolan", Movies: 2},
		}, nil)

		body := getBody(t, setupStatsApp(NewStatsHandler(repo)), "/stats/actor-director-pairs")

		assert.Contains(t, body, `href="/person/christian-bale-1"`)
		assert.Contains(t, body, `href="/person/christopher-nolan-3"`)
	})

	t.Run("shows an empty state without pairs", func(t *testing.T) {
		repo := mocks.NewMockStatsQuerier(t)
		repo.On("GetActorDirectorPairs", "user-123").Return([]types.PersonPair{}, nil)

		body := getBody(t, setupStatsApp(NewStatsHandler(repo)), "/stats/actor-director-pairs")

		assert.Contains(t, body, "No actor and director with more than one movie together")
	})
}
