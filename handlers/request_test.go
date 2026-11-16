package handlers

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

func TestRequestPageBelowOne(t *testing.T) {
	app := fiber.New()

	app.Get("/test-page", func(c *fiber.Ctx) error {
		req := NewRequest(c)

		assert.Equal(t, 1, req.Page())
		assert.Equal(t, 0, req.Offset())

		return c.SendStatus(http.StatusOK)
	})

	for _, page := range []string{"0", "-3"} {
		t.Run(page, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test-page?page="+page, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestRequestAccepts(t *testing.T) {
	app := fiber.New()

	app.Get("/test-accepts", func(c *fiber.Ctx) error {
		req := NewRequest(c)

		assert.True(t, req.Accepts("text/plain"))
		assert.False(t, req.Accepts("application/json"))

		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test-accepts", nil)
	req.Header.Set("Accept", "text/plain; charset=utf-8")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestIsJSON(t *testing.T) {
	tests := []struct {
		accept string
		want   bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"text/html, application/json;q=0.9", true},
		{"Application/JSON", true},
		{"", false},
		{"*/*", false},
		{"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8", false},
	}

	for _, tc := range tests {
		t.Run(tc.accept, func(t *testing.T) {
			app := fiber.New()

			app.Get("/test-json", func(c *fiber.Ctx) error {
				assert.Equal(t, tc.want, NewRequest(c).IsJSON())
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test-json", nil)
			req.Header.Set("Accept", tc.accept)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}
