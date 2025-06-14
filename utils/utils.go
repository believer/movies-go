package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

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

func SelfHealingUrl(text string) string {
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

// Custom NullString and NullInt64 to support parsing
// JSONB from Postgres
type NullString struct {
	sql.NullString
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	var s *string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s != nil {
		ns.String = *s
		ns.Valid = true
	} else {
		ns.String = ""
		ns.Valid = false
	}
	return nil
}

type NullInt64 struct {
	sql.NullInt64
}

func (ni *NullInt64) UnmarshalJSON(b []byte) error {
	var i *int64
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	if i != nil {
		ni.Int64 = *i
		ni.Valid = true
	} else {
		ni.Int64 = 0
		ni.Valid = false
	}
	return nil
}
