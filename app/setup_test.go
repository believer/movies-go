package app

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// A handler's error text can hold SQL, internal paths or secrets. The client
// gets a generic 500 body; the error only reaches the server log.
func TestErrorHandlerDoesNotLeakHandlerErrors(t *testing.T) {
	app := newApp()
	app.Get("/boom", func(c *fiber.Ctx) error {
		return errors.New("sentinel: postgres://admin:password@localhost/movies")
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/boom", nil))
	assert.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	assert.NotContains(t, string(body), "sentinel")
	assert.Equal(t, "Internal Server Error", string(body))
}

// Fiber errors describe a deliberate response, so they keep their status and
// message.
func TestErrorHandlerKeepsFiberErrors(t *testing.T) {
	app := newApp()
	app.Get("/missing", func(c *fiber.Ctx) error {
		return fiber.ErrNotFound
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/missing", nil))
	assert.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "Not Found", string(body))
}
