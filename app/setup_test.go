package app

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v5"
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

func newAuthApp(t *testing.T) *fiber.App {
	t.Helper()
	t.Setenv("ADMIN_SECRET", "test-secret")

	app := newApp()
	app.Use(recover.New())
	app.Use(authenticate)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString(fmt.Sprintf("%v %v", c.Locals("IsAuthenticated"), c.Locals("UserId")))
	})

	return app
}

func signToken(t *testing.T, claims jwt.MapClaims, secret string) string {
	t.Helper()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	assert.NoError(t, err)

	return token
}

func requestWithToken(t *testing.T, app *fiber.App, token string) (*http.Response, string) {
	t.Helper()

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: token})

	resp, err := app.Test(req)
	assert.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	return resp, string(body)
}

func TestAuthenticateAcceptsValidToken(t *testing.T) {
	app := newAuthApp(t)
	token := signToken(t, jwt.MapClaims{
		"id":  "42",
		"exp": time.Now().Add(time.Hour).Unix(),
	}, "test-secret")

	resp, body := requestWithToken(t, app, token)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.Equal(t, "true 42", body)
	assert.Empty(t, resp.Header.Get(fiber.HeaderSetCookie))
}

// A token that doesn't validate counts as logged out, and its cookie is
// cleared so the browser stops sending it
func TestAuthenticateTreatsInvalidTokenAsLoggedOut(t *testing.T) {
	tests := []struct {
		name   string
		claims jwt.MapClaims
		secret string
	}{
		{"expired", jwt.MapClaims{"id": "42", "exp": time.Now().Add(-time.Hour).Unix()}, "test-secret"},
		{"signed with another secret", jwt.MapClaims{"id": "42"}, "other-secret"},
		{"missing id claim", jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}, "test-secret"},
		{"non-string id claim", jwt.MapClaims{"id": 42}, "test-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newAuthApp(t)
			resp, body := requestWithToken(t, app, signToken(t, tt.claims, tt.secret))

			assert.Equal(t, fiber.StatusOK, resp.StatusCode)
			assert.Equal(t, "false 1", body)
			assert.Contains(t, resp.Header.Get(fiber.HeaderSetCookie), "token=;")
		})
	}

	t.Run("malformed", func(t *testing.T) {
		app := newAuthApp(t)
		resp, body := requestWithToken(t, app, "not-a-token")

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		assert.Equal(t, "false 1", body)
		assert.Contains(t, resp.Header.Get(fiber.HeaderSetCookie), "token=;")
	})
}
