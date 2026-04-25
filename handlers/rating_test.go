package handlers

import (
	"believer/movies/mocks"
	"believer/movies/types"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type ratingsHandlerTest struct {
	name         string
	url          string
	mockSetup    func(*mocks.MockRatingsQuerier)
	expectedCode int
}

func setupRatingsApp(h *RatingsHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/rating/:rating", h.GetMoviesByRating)

	return app
}

func runRatingHandlerTests(t *testing.T, tests []ratingsHandlerTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockRatingsQuerier(t)
			tt.mockSetup(repo)

			app := setupRatingsApp(NewRatingsHandler(repo))
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetMoviesByRating(t *testing.T) {
	movies := types.Movies{{Title: "Everything Everywhere All at Once"}}

	runRatingHandlerTests(t, []ratingsHandlerTest{
		{
			name: "returns first 50 by default (0 offset)",
			url:  "/rating/10",
			mockSetup: func(m *mocks.MockRatingsQuerier) {
				m.On("GetMoviesByRating", "user-123", 10, 0).Return(movies, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns for page 4",
			url:  "/rating/8?page=4",
			mockSetup: func(m *mocks.MockRatingsQuerier) {
				m.On("GetMoviesByRating", "user-123", 8, 150).Return(movies, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns empty OK when no more movies",
			url:  "/rating/8?page=4",
			mockSetup: func(m *mocks.MockRatingsQuerier) {
				m.On("GetMoviesByRating", "user-123", 8, 150).Return(types.Movies{}, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns error if rating cannot be parsed",
			url:          "/rating/invalid",
			mockSetup:    func(m *mocks.MockRatingsQuerier) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "returns error if movie query breaks",
			url:  "/rating/8",
			mockSetup: func(m *mocks.MockRatingsQuerier) {
				m.On("GetMoviesByRating", "user-123", 8, 0).Return(types.Movies{}, fmt.Errorf("SQL error"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	})
}
