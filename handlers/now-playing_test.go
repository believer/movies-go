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

type nowPlayingHandlerTest struct {
	name         string
	url          string
	mockSetup    func(*mocks.MockNowPlayingQuerier)
	expectedCode int
}

func setupNowPlayingApp(h *NowPlayingHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/now-playing", h.GetNowPlaying)

	return app
}

func runNowPlayingHandlerTests(t *testing.T, tests []nowPlayingHandlerTest) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockNowPlayingQuerier(t)
			tt.mockSetup(repo)

			app := setupNowPlayingApp(NewNowPlayingHandler(repo))
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.StatusCode)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetNowPlaying(t *testing.T) {
	movies := types.Movies{{Title: "Everything Everywhere All at Once"}}

	runNowPlayingHandlerTests(t, []nowPlayingHandlerTest{
		{
			name: "returns now playing movies",
			url:  "/now-playing",
			mockSetup: func(m *mocks.MockNowPlayingQuerier) {
				m.On("GetNowPlaying", "user-123").Return(movies, nil)
			},
			expectedCode: http.StatusOK,
		},
	})
}
