package handlers

import (
	"believer/movies/mocks"
	"believer/movies/types"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupNowPlayingApp(h *NowPlayingHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", c.Cookies("token") != "")
		return c.Next()
	})

	app.Get("/now-playing", h.GetNowPlaying)
	app.Get("/now-playing/manage", h.GetManage)
	app.Post("/now-playing/manage", h.PostManage)
	app.Put("/now-playing/:id", h.PutNowPlaying)
	app.Delete("/now-playing/:id", h.DeleteNowPlaying)

	return app
}

func TestNowPlayingHandlerGet(t *testing.T) {
	movies := types.Movies{{Title: "Everything Everywhere All at Once"}}

	t.Run("returns now playing movies", func(t *testing.T) {
		repo := mocks.NewMockNowPlayingQuerier(t)
		movieRepo := mocks.NewMockMovieQuerier(t)
		repo.On("GetNowPlaying", "user-123").Return(movies, nil)

		app := setupNowPlayingApp(NewNowPlayingHandler(repo, movieRepo))
		req := httptest.NewRequest(http.MethodGet, "/now-playing", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestNowPlayingHandlerGetManage(t *testing.T) {
	movies := types.Movies{{Title: "Everything Everywhere All at Once"}}

	t.Run("redirects to login when unauthenticated", func(t *testing.T) {
		repo := mocks.NewMockNowPlayingQuerier(t)
		movieRepo := mocks.NewMockMovieQuerier(t)
		app := setupNowPlayingApp(NewNowPlayingHandler(repo, movieRepo))

		req := httptest.NewRequest(http.MethodGet, "/now-playing/manage", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/login", resp.Header.Get("Location"))
	})

	t.Run("returns 200 with manage view when authenticated", func(t *testing.T) {
		repo := mocks.NewMockNowPlayingQuerier(t)
		movieRepo := mocks.NewMockMovieQuerier(t)
		repo.On("GetNowPlaying", "user-123").Return(movies, nil)

		app := setupNowPlayingApp(NewNowPlayingHandler(repo, movieRepo))
		req := httptest.NewRequest(http.MethodGet, "/now-playing/manage", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestNowPlayingHandlerPostManage(t *testing.T) {
	t.Run("adds a movie and redirects", func(t *testing.T) {
		repo := mocks.NewMockNowPlayingQuerier(t)
		movieRepo := mocks.NewMockMovieQuerier(t)

		movieRepo.On("MovieExists", "tt0111161").Return(true, nil)
		movieRepo.On("GetMovieByImdbID", "tt0111161").Return(types.Movie{ID: 1, Title: "The Shawshank Redemption"}, nil)
		movieRepo.On("UpdateNowPlaying", 1, 10.0, "user-123").Return(nil)

		app := setupNowPlayingApp(NewNowPlayingHandler(repo, movieRepo))
		
		form := url.Values{}
		form.Add("imdb_id", "tt0111161")
		form.Add("position", "10")

		req := httptest.NewRequest(http.MethodPost, "/now-playing/manage", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})
		
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, "/now-playing/manage", resp.Header.Get("Location"))
	})
}

func TestNowPlayingHandlerPut(t *testing.T) {
	t.Run("updates now playing and returns updated item", func(t *testing.T) {
		repo := mocks.NewMockNowPlayingQuerier(t)
		movieRepo := mocks.NewMockMovieQuerier(t)

		movieRepo.On("UpdateNowPlaying", 1, 45.0, "user-123").Return(nil)
		movieRepo.On("GetMovieByIDSimple", 1).Return(types.Movie{ID: 1, Title: "The Shawshank Redemption", Runtime: 142}, nil)

		app := setupNowPlayingApp(NewNowPlayingHandler(repo, movieRepo))

		form := url.Values{}
		form.Add("position", "45")

		req := httptest.NewRequest(http.MethodPut, "/now-playing/1", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestNowPlayingHandlerDelete(t *testing.T) {
	t.Run("deletes now playing", func(t *testing.T) {
		repo := mocks.NewMockNowPlayingQuerier(t)
		movieRepo := mocks.NewMockMovieQuerier(t)

		movieRepo.On("DeleteNowPlayingDirect", "user-123", 1).Return(nil)

		app := setupNowPlayingApp(NewNowPlayingHandler(repo, movieRepo))

		req := httptest.NewRequest(http.MethodDelete, "/now-playing/1", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})

		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
