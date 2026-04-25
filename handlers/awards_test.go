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

type awardsHandlerTest struct {
	name         string
	url          string
	mockSetup    func(*mocks.MockAwardsQuerier)
	expectedCode int
}

func setupAwardsApp(h *AwardsHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/awards/:awards", h.GetMoviesByNumberOfAwards)
	app.Get("/awards/year/:year", h.GetAwardsByYear)

	return app
}

func runAwardsHandlerTests(t *testing.T, tests []awardsHandlerTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockAwardsQuerier(t)
			tt.mockSetup(repo)

			app := setupAwardsApp(NewAwardsHandler(repo))
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetMoviesByNumberOfAwards(t *testing.T) {
	movies := types.Movies{{Title: "Everything Everywhere All at Once"}}

	runAwardsHandlerTests(t, []awardsHandlerTest{
		{
			name: "returns wins by default",
			url:  "/awards/2?type=academy-award",
			mockSetup: func(m *mocks.MockAwardsQuerier) {
				m.On("GetByWins", "user-123", 2, "academy-award").Return(movies, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns nominations when flag is set",
			url:  "/awards/2?type=academy-award&nominations=true",
			mockSetup: func(m *mocks.MockAwardsQuerier) {
				m.On("GetByNominations", "user-123", 2, "academy-award").Return(movies, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "rejects missing award type",
			url:          "/awards/2",
			mockSetup:    func(m *mocks.MockAwardsQuerier) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "rejects unknown award type",
			url:          "/awards/2?type=grammy",
			mockSetup:    func(m *mocks.MockAwardsQuerier) {},
			expectedCode: http.StatusBadRequest,
		},
	})
}

func TestGetAwardsByYear(t *testing.T) {
	moviesGrouped := []types.AwardsByYear{{Title: "Titanic"}}
	categoriesGrouped := []types.AwardsByCategory{{Category: "Test"}}

	runAwardsHandlerTests(t, []awardsHandlerTest{
		{
			name: "returns sort by movies by default",
			url:  "/awards/year/2026?type=academy-award",
			mockSetup: func(m *mocks.MockAwardsQuerier) {
				m.On("GetGroupedByMovie", "2026", "academy-award").Return(moviesGrouped, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns sort by movies",
			url:  "/awards/year/2026?type=academy-award&sort=Movie",
			mockSetup: func(m *mocks.MockAwardsQuerier) {
				m.On("GetGroupedByMovie", "2026", "academy-award").Return(moviesGrouped, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "returns sort by categories",
			url:  "/awards/year/2026?type=academy-award&sort=Category",
			mockSetup: func(m *mocks.MockAwardsQuerier) {
				m.On("GetGroupedByCategory", "2026", "academy-award").Return(categoriesGrouped, nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:         "rejects missing award type",
			url:          "/awards/year/2026",
			mockSetup:    func(m *mocks.MockAwardsQuerier) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "rejects unknown award type",
			url:          "/awards/year/2026?type=grammy",
			mockSetup:    func(m *mocks.MockAwardsQuerier) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "rejects unknown sort category",
			url:          "/awards/year/2026?type=academy-award&sort=INVALID",
			mockSetup:    func(m *mocks.MockAwardsQuerier) {},
			expectedCode: http.StatusBadRequest,
		},
	})
}
