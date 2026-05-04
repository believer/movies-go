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

type yearsHandlerTest struct {
	name         string
	url          string
	mockSetup    func(*mocks.MockYearsQuerier)
	expectedCode int
}

func setupYearsApp(h *YearsHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/year/:year", h.GetMoviesByYear)

	return app
}

func runYearsHandlerTests(t *testing.T, tests []yearsHandlerTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockYearsQuerier(t)
			tt.mockSetup(repo)

			app := setupYearsApp(NewYearsHandler(repo))
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetMoviesByYears(t *testing.T) {
	movies := types.Movies{{Title: "Everything Everywhere All at Once"}}

	runYearsHandlerTests(t, []yearsHandlerTest{
		{
			name: "returns movies by default",
			url:  "/year/2026",
			mockSetup: func(m *mocks.MockYearsQuerier) {
				m.On("GetMoviesByYear", "user-123", "2026", 0).Return(movies, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns error if request fails",
			url:  "/year/2026",
			mockSetup: func(m *mocks.MockYearsQuerier) {
				m.On("GetMoviesByYear", "user-123", "2026", 0).Return(movies, fmt.Errorf("Test"))
			},
			expectedCode: http.StatusInternalServerError,
		},
	})
}
