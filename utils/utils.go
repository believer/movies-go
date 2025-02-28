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
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	imdbPattern = regexp.MustCompile(`^tt\d{7,}$`)
	tmdbPattern = regexp.MustCompile(`^\d+$`)
)

func IsAuthenticated(c *fiber.Ctx) bool {
	return c.Cookies("token") != ""
}

func ParseId(s string) (string, error) {
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

	if len(parts) == 0 {
		return "0"
	}

	return strings.Join(parts, " ")
}

func TemplRender(c *fiber.Ctx, component templ.Component) error {
	return adaptor.HTTPHandler(templ.Handler(component))(c)
}

func Slugify(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Replace spaces with hyphens
	text = strings.ReplaceAll(text, " ", "-")

	// Remove special characters
	re := regexp.MustCompile(`[^a-z0-9\-]`)
	text = re.ReplaceAllString(text, "")

	// Replace multiple hyphens with a single one
	re = regexp.MustCompile(`-+`)
	text = re.ReplaceAllString(text, "-")

	// Trim hyphens from the start and end
	text = strings.Trim(text, "-")

	return text
}

func SelfHealingUrl(text string) string {
	parts := strings.Split(path.Base(text), "-")

	return parts[len(parts)-1]
}

func Formatter() *message.Printer {
	return message.NewPrinter(language.English)
}
