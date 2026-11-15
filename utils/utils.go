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

	slugInvalid = regexp.MustCompile(`[^a-z0-9\-]`)
	slugHyphens = regexp.MustCompile(`-+`)
)

// ParseID returns the IMDb ID (tt0111161) or TMDB ID (27205) in s, which can
// be the ID itself or a URL ending in it.
func ParseID(s string) (string, error) {
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
	text = slugInvalid.ReplaceAllString(text, "")

	// Replace multiple hyphens with a single one
	text = slugHyphens.ReplaceAllString(text, "-")

	// Trim hyphens from the start and end
	text = strings.Trim(text, "-")

	return text
}

// CreateSelfHealingURL builds /root/<slug of name>-id. Only the ID after
// the last hyphen identifies the page, so a renamed item's old links still
// work.
func CreateSelfHealingURL(root, name, id string) string {
	return fmt.Sprintf("/%s/%s-%s", root, Slugify(name), id)
}

// SelfHealingID returns the ID after the last hyphen of a self-healing URL.
// IDs that aren't numbers, like a production country's "US", are returned
// as they are.
func SelfHealingID(text string) string {
	base := path.Base(text)

	return base[strings.LastIndex(base, "-")+1:]
}

// SelfHealingNumericID returns the ID of a self-healing URL, or an error when
// the ID isn't a number.
func SelfHealingNumericID(text string) (string, error) {
	id := SelfHealingID(text)

	if _, err := strconv.Atoi(id); err != nil {
		return "", errors.New("not a valid ID")
	}

	return id, nil
}

var englishPrinter = message.NewPrinter(language.English)

// FormatNumber formats n with thousands separators, e.g. 12,345
func FormatNumber(n int) string {
	return englishPrinter.Sprintf("%d", n)
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

func AvailableYears(now time.Time) []string {
	// First year with "real" data
	// 2011 is used as a catch all for anything before I had the database
	endYear := 2012
	years := make([]string, 0)

	for year := now.Year(); year >= endYear; year-- {
		y := strconv.Itoa(year)
		years = append(years, y)
	}

	return years
}

// ParseWatchedAt parses the watched_at form value, either a date and time
// ("2006-01-02T15:04") or a date. A date gets the time of day from now, and
// anything else falls back to now.
func ParseWatchedAt(value string, now time.Time) time.Time {
	if t, err := time.Parse("2006-01-02T15:04", value); err == nil {
		return t
	}

	if d, err := time.Parse("2006-01-02", value); err == nil {
		return time.Date(d.Year(), d.Month(), d.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())
	}

	return now
}
