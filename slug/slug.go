package slug

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/unicode/norm"
)

var (
	slugInvalid = regexp.MustCompile(`[^a-z0-9\-]`)
	slugHyphens = regexp.MustCompile(`-+`)
)

func Make(text string) string {
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

// URL builds /root/<slug of name>-id. Only the ID after
// the last hyphen identifies the page, so a renamed item's old links still
// work.
func URL(root, name, id string) string {
	return fmt.Sprintf("/%s/%s-%s", root, Make(name), id)
}

// ID returns the ID after the last hyphen of a self-healing URL.
// IDs that aren't numbers, like a production country's "US", are returned
// as they are.
func ID(text string) string {
	base := path.Base(text)

	return base[strings.LastIndex(base, "-")+1:]
}

// NumericID returns the ID of a self-healing URL, or an error when
// the ID isn't a number.
func NumericID(text string) (string, error) {
	id := ID(text)

	if _, err := strconv.Atoi(id); err != nil {
		return "", errors.New("not a valid ID")
	}

	return id, nil
}
