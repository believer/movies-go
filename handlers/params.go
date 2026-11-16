package handlers

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	imdbPattern = regexp.MustCompile(`^tt\d{7,}$`)
	tmdbPattern = regexp.MustCompile(`^\d+$`)
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
