package utils

import (
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
	return r.Cookies("token") != ""
}

// MovieID returns the self-healed movie ID or "0" if parsing fails.
func (r Request) MovieID() string {
	id, err := SelfHealingUrl(r.Params("id"))
	if err != nil {
		return "0"
	}
	return id
}

// IDString returns the self-healed ID from path as a string.
func (r Request) IDString() string {
	return SelfHealingUrlString(r.Params("id"))
}

// Year returns the "year" query parameter, defaulting to "All".
func (r Request) Year() string {
	return r.QueryDefault("year", "All")
}

// Page returns the "page" query parameter, defaulting to 1.
func (r Request) Page() int {
	return r.QueryInt("page", 1)
}

// Offset returns the SQL query offset based on the current page.
func (r Request) Offset() int {
	return (r.Page() - 1) * 50
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
