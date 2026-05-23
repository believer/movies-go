package handlers

import (
	"believer/movies/mocks"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupSettingsApp(h *SettingsHandler) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/settings", h.Settings)
	app.Put("/settings/watch-providers", h.SettingsWatchProviders)

	return app
}

func TestSettings(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		repo := mocks.NewMockSettingsQuerier(t)
		app := setupSettingsApp(NewSettingsHandler(repo))

		req := httptest.NewRequest(http.MethodGet, "/settings", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorized", func(t *testing.T) {
		repo := mocks.NewMockSettingsQuerier(t)
		repo.On("GetWatchProviders", "user-123").Return("Netflix,HBO Max", nil)

		app := setupSettingsApp(NewSettingsHandler(repo))

		req := httptest.NewRequest(http.MethodGet, "/settings", nil)
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestSettingsWatchProviders(t *testing.T) {
	t.Run("unauthorized", func(t *testing.T) {
		repo := mocks.NewMockSettingsQuerier(t)
		app := setupSettingsApp(NewSettingsHandler(repo))

		req := httptest.NewRequest(http.MethodPut, "/settings/watch-providers", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("authorized", func(t *testing.T) {
		repo := mocks.NewMockSettingsQuerier(t)
		repo.On("UpdateWatchProviders", "user-123", "Netflix,HBO Max").Return(nil)

		app := setupSettingsApp(NewSettingsHandler(repo))

		form := url.Values{}
		form.Add("providers", "Netflix")
		form.Add("providers", "HBO Max")

		req := httptest.NewRequest(http.MethodPut, "/settings/watch-providers", bytes.NewBufferString(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "token", Value: "active"})

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
