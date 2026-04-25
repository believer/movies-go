package handlers

import (
	"believer/movies/mocks"
	"believer/movies/types"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type seriesHandlerTest struct {
	name         string
	url          string
	mockSetup    func(*mocks.MockSeriesQuerier)
	mockAssert   func(*mocks.MockSeriesQuerier)
	expectedCode int
}

func setupSeriesApp(h *SeriesHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/series/:id", h.GetSeries)

	return app
}

func runSeriesHandlerTests(t *testing.T, tests []seriesHandlerTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockSeriesQuerier(t)
			tt.mockSetup(repo)

			app := setupSeriesApp(NewSeriesHandler(repo))
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)

			if tt.mockAssert != nil {
				tt.mockAssert(repo)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestGetSeries(t *testing.T) {
	series := types.Series{Name: "Alien"}
	movies := []types.SeriesMovies{
		{
			Name:   "Alien Anthology",
			Movies: types.MoviesInSeries{{Title: "Aliens"}},
		},
	}

	runSeriesHandlerTests(t, []seriesHandlerTest{
		{
			name: "returns series and movies",
			url:  "/series/test-1",
			mockSetup: func(m *mocks.MockSeriesQuerier) {
				m.On("GetSeriesByID", "1").Return(series, nil)
				m.On("GetSeriesMovies", "1", "user-123").Return(movies, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns 500 for series error",
			url:  "/series/test-1",
			mockSetup: func(m *mocks.MockSeriesQuerier) {
				m.On("GetSeriesByID", "1").Return(series, fmt.Errorf("Error"))
			},
			mockAssert: func(m *mocks.MockSeriesQuerier) {
				m.AssertNotCalled(t, "GetSeriesMovies", "1", "user-123")
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "returns 500 for series movies error",
			url:  "/series/test-1",
			mockSetup: func(m *mocks.MockSeriesQuerier) {
				m.On("GetSeriesByID", "1").Return(series, nil)
				m.On("GetSeriesMovies", "1", "user-123").Return(movies, fmt.Errorf("Error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name: "passes with no rows errors",
			url:  "/series/test-1",
			mockSetup: func(m *mocks.MockSeriesQuerier) {
				m.On("GetSeriesByID", "1").Return(series, sql.ErrNoRows)
				m.On("GetSeriesMovies", "1", "user-123").Return(movies, sql.ErrNoRows)
			},
			expectedCode: http.StatusOK,
		},
	})
}

func TestCalculateSeriesStats(t *testing.T) {
	tests := []struct {
		name          string
		movies        []types.SeriesMovies
		expectedTotal int
		expectedSeen  []int // seen count per series group
	}{
		{
			name: "counts total movies across series",
			movies: []types.SeriesMovies{
				{Movies: types.MoviesInSeries{{Title: "Alien", Seen: true}, {Title: "Aliens", Seen: false}}},
				{Movies: types.MoviesInSeries{{Title: "Alien 3", Seen: true}}},
			},
			expectedTotal: 3,
			expectedSeen:  []int{1, 1},
		},
		{
			name:          "empty series",
			movies:        []types.SeriesMovies{},
			expectedTotal: 0,
			expectedSeen:  []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, result := calculateSeriesStats(tt.movies)
			assert.Equal(t, tt.expectedTotal, total)

			for i, s := range result {
				assert.Equal(t, tt.expectedSeen[i], s.Seen)
			}
		})
	}
}
