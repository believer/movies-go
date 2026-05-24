package handlers

import (
	"believer/movies/mocks"
	"believer/movies/types"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupReviewApp(h *ReviewHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", c.Cookies("token") != "")
		return c.Next()
	})

	app.Get("/review/new", h.AddMovieReview)
	app.Post("/review/new", h.InsertMovieReview)
	app.Get("/review/:id/edit", h.EditMovieReview)
	app.Put("/review/:id", h.UpdateMovieReview)
	app.Delete("/review/:id", h.DeleteMovieReview)

	return app
}

func TestAddMovieReview(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		app := setupReviewApp(NewReviewHandler(repo))

		req := httptest.NewRequest(http.MethodGet, "/review/new?movieId=456", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		app := setupReviewApp(NewReviewHandler(repo))

		req := httptest.NewRequest(http.MethodGet, "/review/new?movieId=456", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestInsertMovieReview(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		app := setupReviewApp(NewReviewHandler(repo))

		req := httptest.NewRequest(http.MethodPost, "/review/new?movieId=456", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		mockReview := types.Review{
			ID:      1,
			Content: "Great movie",
			Private: false,
		}
		repo.On("InsertReview", "Great movie", false, "user-123", 456).Return(mockReview, nil)

		app := setupReviewApp(NewReviewHandler(repo))

		form := url.Values{}
		form.Add("review", "Great movie")
		form.Add("review_private", "false")

		req := httptest.NewRequest(http.MethodPost, "/review/new?movieId=456", bytes.NewBufferString(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestEditMovieReview(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		app := setupReviewApp(NewReviewHandler(repo))

		req := httptest.NewRequest(http.MethodGet, "/review/1/edit?movieId=456", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		mockReview := types.Review{
			ID:      1,
			Content: "Great movie",
			Private: false,
		}
		repo.On("GetReviewByID", "1").Return(mockReview, nil)

		app := setupReviewApp(NewReviewHandler(repo))

		req := httptest.NewRequest(http.MethodGet, "/review/1/edit?movieId=456", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestUpdateMovieReview(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		app := setupReviewApp(NewReviewHandler(repo))

		req := httptest.NewRequest(http.MethodPut, "/review/1?movieId=456", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		mockReview := types.Review{
			ID:      1,
			Content: "Updated content",
			Private: true,
		}
		repo.On("UpdateReview", "1", "Updated content", true).Return(mockReview, nil)

		app := setupReviewApp(NewReviewHandler(repo))

		form := url.Values{}
		form.Add("review", "Updated content")
		form.Add("review_private", "true")

		req := httptest.NewRequest(http.MethodPut, "/review/1?movieId=456", bytes.NewBufferString(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestDeleteMovieReview(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		app := setupReviewApp(NewReviewHandler(repo))

		req := httptest.NewRequest(http.MethodDelete, "/review/1?movieId=456", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorized", func(t *testing.T) {
		repo := mocks.NewMockReviewQuerier(t)
		repo.On("DeleteReview", "1").Return(nil)

		app := setupReviewApp(NewReviewHandler(repo))

		req := httptest.NewRequest(http.MethodDelete, "/review/1?movieId=456", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
