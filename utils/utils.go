package utils

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
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