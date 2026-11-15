package utils

import (
	"mime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Request struct {
	*fiber.Ctx
}

func NewRequest(c *fiber.Ctx) Request {
	return Request{c}
}

// UserID returns the authenticated user's ID or an empty string.
func (r Request) UserID() string {
	if val := r.Locals("UserId"); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// IsAuthenticated returns true if the user is authenticated, otherwise false.
func (r Request) IsAuthenticated() bool {
	if val := r.Locals("IsAuthenticated"); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// MovieID returns the self-healed movie ID or "0" if parsing fails.
func (r Request) MovieID() string {
	id, err := SelfHealingNumericID(r.Params("id"))
	if err != nil {
		return "0"
	}
	return id
}

// IDString returns the self-healed ID from path as a string.
func (r Request) IDString() string {
	return SelfHealingID(r.Params("id"))
}

// Year returns the "year" query parameter, defaulting to "All".
func (r Request) Year() string {
	return r.QueryDefault("year", "All")
}

// pageSize matches the LIMIT 50 in the paged queries in db/
const pageSize = 50

// Page returns the "page" query parameter, defaulting to 1. A page below 1
// is page 1, so Offset never goes negative.
func (r Request) Page() int {
	return max(r.QueryInt("page", 1), 1)
}

// Offset returns the SQL query offset based on the current page.
func (r Request) Offset() int {
	return (r.Page() - 1) * pageSize
}

// AvailableYears returns the list of years for stats and filters.
func (r Request) AvailableYears() []string {
	return append([]string{"All"}, AvailableYears(time.Now())...)
}

// QueryDefault returns a query parameter, falling back to defaultValue if empty.
func (r Request) QueryDefault(key, defaultValue string) string {
	if val := r.Query(key); val != "" {
		return val
	}
	return defaultValue
}

// IsJSON reports whether application/json is one of the media types in
// Accept, with or without parameters. A wildcard like */* doesn't count, so
// a browser still gets HTML.
func (r Request) IsJSON() bool {
	for _, accepted := range strings.Split(r.Get("Accept"), ",") {
		mediaType, _, err := mime.ParseMediaType(accepted)

		if err == nil && mediaType == "application/json" {
			return true
		}
	}

	return false
}
