package utils

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

var (
	imdbPattern = regexp.MustCompile(`^tt\d{7,}$`)
	tmdbPattern = regexp.MustCompile(`^\d+$`)
)

func IsAuthenticated(c *fiber.Ctx) bool {
	cookieAdminSecret := c.Cookies("token")

	return cookieAdminSecret != ""
}

func ParseImdbId(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("Empty ID")
	}

	parsedUrl, err := url.Parse(s)

	if err != nil {
		return "", err
	}

	id := strings.ToLower(strings.TrimRight(path.Base(parsedUrl.Path), "/"))

	if id == "" {
		return "", fmt.Errorf("Empty ID")
	}

	if imdbPattern.MatchString(id) || tmdbPattern.MatchString(id) {
		return id, nil
	}

	return "", fmt.Errorf("Invalid ID format: %s", id)
}

func FormatRuntime(runtime int) string {
	days := runtime / 1440
	hours := runtime / 60 % 24
	minutes := runtime % 60

	parts := []string{}

	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}

	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	return strings.Join(parts, " ")
}

func TemplRender(c *fiber.Ctx, component templ.Component) error {
	return adaptor.HTTPHandler(templ.Handler(component))(c)
}
