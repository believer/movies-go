package utils

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func IsAuthenticated(c *fiber.Ctx) bool {
	adminSecret := os.Getenv("ADMIN_SECRET")
	cookieAdminSecret := c.Cookies("admin_secret")

	return cookieAdminSecret == adminSecret
}

func ParseImdbId(s string) (string, error) {
	if s == "" {
		return "", fmt.Errorf("Empty IMDb ID")
	}

	parsedUrl, err := url.Parse(s)

	if err != nil {
		return "", err
	}

	imdbId := path.Base(parsedUrl.Path)
	imdbId = strings.TrimRight(imdbId, "/")
	imdbId = strings.ToLower(imdbId)

	if imdbId == "" {
		return "", fmt.Errorf("Empty IMDb ID")
	}

	// IMDb IDs start with "tt" followed by 7 or more digits
	match, _ := regexp.MatchString(`^tt\d{7,}$`, imdbId)

	if !match {
		return "", fmt.Errorf("Invalid IMDb ID format: %s", imdbId)
	}

	return imdbId, nil
}

func FormatRuntime(runtime int) string {
	var dayStr, hourStr, minStr string
	var (
		days    = runtime / 1440
		hours   = runtime / 60 % 24
		minutes = runtime % 60
	)

	if days != 0 {
		dayStr = fmt.Sprintf("%dd", days)

		if hours != 0 || minutes != 0 {
			dayStr += " "
		}
	}

	if hours != 0 {
		hourStr = fmt.Sprintf("%dh", hours)

		if minutes != 0 {
			hourStr += " "
		}
	}

	if minutes != 0 {
		minStr = fmt.Sprintf("%dm", minutes)
	}

	return dayStr + hourStr + minStr
}

func TemplRender(c *fiber.Ctx, component templ.Component) error {
	return adaptor.HTTPHandler(templ.Handler(component))(c)
}
