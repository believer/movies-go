package router

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/stretchr/testify/assert"
)

// publicRoutes are the routes anonymous users may reach. Every other route
// must reject an anonymous request. Making a route public means adding it
// here.
var publicRoutes = map[string]bool{
	"GET /health": true,
	"GET /":       true,

	"GET /login":   true,
	"POST /login":  true,
	"POST /logout": true,

	"GET /watchlist/":                  true,
	"GET /watchlist/movies":            true,
	"GET /watchlist/unreleased-movies": true,

	"GET /movie/":                true,
	"GET /movie/imdb":            true,
	"GET /movie/:id":             true,
	"GET /movie/:imdbId/awards":  true,
	"GET /movie/:id/seen/others": true,
	"GET /movie/:id/rating":      true,
	"GET /movie/:id/lists":       true,

	"GET /year/":      true,
	"GET /year/:year": true,

	"GET /person/":    true,
	"GET /person/:id": true,

	"GET /genre/":      true,
	"GET /genre/stats": true,
	"GET /genre/:id":   true,

	"GET /language/":      true,
	"GET /language/stats": true,
	"GET /language/:id":   true,

	"GET /production-company/":      true,
	"GET /production-company/stats": true,
	"GET /production-company/:id":   true,

	"GET /production-country/":      true,
	"GET /production-country/stats": true,
	"GET /production-country/:id":   true,

	"GET /awards/":           true,
	"GET /awards/:awards":    true,
	"GET /awards/year/:year": true,

	"GET /rating/":        true,
	"GET /rating/:rating": true,

	"GET /stats/":                         true,
	"GET /stats/ratings":                  true,
	"GET /stats/wilhelm-scream":           true,
	"GET /stats/ratings/:year":            true,
	"GET /stats/by-month":                 true,
	"GET /stats/by-weekday":               true,
	"GET /stats/movies-by-year":           true,
	"GET /stats/highest-ranked-person":    true,
	"GET /stats/best-of-the-year":         true,
	"GET /stats/most-watched-person/:job": true,
	"GET /stats/seen-with":                true,
	"GET /stats/hot-takes":                true,
	"GET /stats/streaks":                  true,
	"GET /stats/actor-director-pairs":     true,

	"GET /series/:id": true,

	"GET /now-playing/": true,

	"GET /list/":    true,
	"GET /list/:id": true,
}

var pathParam = regexp.MustCompile(`:[^/]+`)

// TestNonPublicRoutesRequireLogin sends every route outside publicRoutes an
// anonymous request and expects a 401, or for pages a redirect to the start
// page or the login page. It runs with a nil db.Client, so a guarded handler must
// reject the request before it touches a repository; one that doesn't panics
// into a 500 and fails the test.
func TestNonPublicRoutesRequireLogin(t *testing.T) {
	app := fiber.New()
	app.Use(recover.New())
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("IsAuthenticated", false)
		c.Locals("UserId", "1")
		return c.Next()
	})

	SetupRoutes(app)

	registered := map[string]bool{}

	for _, route := range app.GetRoutes(true) {
		// Fiber adds a HEAD route for each GET route, served by the same handler
		if route.Method == http.MethodHead {
			continue
		}

		key := route.Method + " " + route.Path
		registered[key] = true

		if publicRoutes[key] {
			continue
		}

		t.Run(key, func(t *testing.T) {
			path := pathParam.ReplaceAllString(route.Path, "1")
			req := httptest.NewRequest(route.Method, path, nil)
			resp, err := app.Test(req)

			assert.NoError(t, err)

			hint := "%s doesn't reject anonymous requests; if it's meant to be public, add it to publicRoutes"

			if resp.StatusCode == fiber.StatusFound {
				assert.Contains(t, []string{"/", "/login"}, resp.Header.Get("Location"), hint, key)
			} else {
				assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode, hint, key)
			}
		})
	}

	for key := range publicRoutes {
		assert.True(t, registered[key], "public route %q is not registered", key)
	}
}
