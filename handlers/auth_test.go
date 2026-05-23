package handlers

import (
	"believer/movies/db"
	"believer/movies/mocks"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthApp(h *AuthHandler) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("IsAuthenticated", false)
		c.Locals("UserId", "")
		return c.Next()
	})
	app.Get("/login", h.GetLogin)
	app.Post("/login", h.Login)
	app.Post("/logout", h.Logout)
	return app
}

func TestGetLogin(t *testing.T) {
	repo := mocks.NewMockAuthQuerier(t)
	app := setupAuthApp(NewAuthHandler(repo))

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLogin(t *testing.T) {
	t.Run("invalid username", func(t *testing.T) {
		repo := mocks.NewMockAuthQuerier(t)
		repo.On("GetUserForLogin", "nonexistent").Return(db.UserAuth{}, assert.AnError)

		app := setupAuthApp(NewAuthHandler(repo))

		form := url.Values{}
		form.Add("username", "nonexistent")
		form.Add("password", "pass")

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid credentials", func(t *testing.T) {
		os.Setenv("ADMIN_SECRET", "super-secret-key")
		defer os.Unsetenv("ADMIN_SECRET")

		hashed, _ := bcrypt.GenerateFromPassword([]byte("secretpassword"), bcrypt.DefaultCost)

		repo := mocks.NewMockAuthQuerier(t)
		repo.On("GetUserForLogin", "validuser").Return(db.UserAuth{
			ID:           "123",
			PasswordHash: string(hashed),
		}, nil)

		app := setupAuthApp(NewAuthHandler(repo))

		form := url.Values{}
		form.Add("username", "validuser")
		form.Add("password", "secretpassword")

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
		assert.Equal(t, "/", resp.Header.Get("Location"))
	})
}

func TestLogout(t *testing.T) {
	repo := mocks.NewMockAuthQuerier(t)
	app := setupAuthApp(NewAuthHandler(repo))

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
