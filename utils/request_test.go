package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRequestHelpersAuthenticated(t *testing.T) {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("UserId", "user-123")
		c.Locals("IsAuthenticated", true)
		return c.Next()
	})

	app.Get("/test", func(c *fiber.Ctx) error {
		req := NewRequest(c)

		assert.Equal(t, "user-123", req.UserID())
		assert.True(t, req.IsAuthenticated())
		assert.True(t, IsAuthenticated(c))
		assert.Equal(t, "2024", req.Year())
		assert.Equal(t, 2, req.Page())
		assert.Equal(t, 50, req.Offset())
		assert.Equal(t, "custom", req.QueryDefault("other", "fallback"))
		assert.Equal(t, "fallback2", req.QueryDefault("missing", "fallback2"))

		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test?year=2024&page=2&other=custom", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestHelpersUnauthenticated(t *testing.T) {
	app := fiber.New()

	app.Get("/test-unauth", func(c *fiber.Ctx) error {
		req := NewRequest(c)

		assert.Equal(t, "", req.UserID())
		assert.False(t, req.IsAuthenticated())
		assert.False(t, IsAuthenticated(c))
		assert.Equal(t, "All", req.Year())
		assert.Equal(t, 1, req.Page())
		assert.Equal(t, 0, req.Offset())

		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test-unauth", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestAvailableYears(t *testing.T) {
	app := fiber.New()

	app.Get("/test-years", func(c *fiber.Ctx) error {
		req := NewRequest(c)
		years := req.AvailableYears()

		assert.NotEmpty(t, years)
		assert.Equal(t, "All", years[0])
		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test-years", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
