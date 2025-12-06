package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/unicode/norm"
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
		return "", fmt.Errorf("empty ID")
	}

	parsedUrl, err := url.Parse(s)

	if err != nil {
		return "", err
	}

	id := strings.ToLower(strings.TrimRight(path.Base(parsedUrl.Path), "/"))

	if id == "" {
		return "", fmt.Errorf("empty ID")
	}

	if imdbPattern.MatchString(id) || tmdbPattern.MatchString(id) {
		return id, nil
	}

	return "", fmt.Errorf("invalid ID format: %s", id)
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
		return "0m"
	}

	return strings.Join(parts, " ")
}

func Render(c *fiber.Ctx, component templ.Component) error {
	c.Set("Content-Type", "text/html")
	return component.Render(c.Context(), c.Response().BodyWriter())
}

func Slugify(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Normalize diacritical marks
	text = norm.NFD.String(text)

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

func SelfHealingUrl(text string) (string, error) {
	parts := strings.Split(path.Base(text), "-")
	id := parts[len(parts)-1]
	_, err := strconv.Atoi(parts[len(parts)-1])

	if err != nil {
		return "", errors.New("not a valid ID")
	}

	return id, nil
}

func SelfHealingUrlString(text string) string {
	parts := strings.Split(path.Base(text), "-")
	return parts[len(parts)-1]
}

func Formatter() *message.Printer {
	return message.NewPrinter(language.English)
}

func ScanJSON[T any](v any, target *T) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, target)
	case string:
		return json.Unmarshal([]byte(vv), target)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

func AvailableYears() []string {
	// First year with "real" data
	// 2011 is used as a catch all for anything before I had the database
	endYear := 2012
	currentYear := time.Now().Year()

	years := make([]string, 0)

	for year := currentYear; year >= endYear; year-- {
		y := strconv.Itoa(year)
		years = append(years, y)
	}

	return years
}
