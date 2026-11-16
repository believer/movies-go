package format

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func Runtime(runtime int) string {
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

var englishPrinter = message.NewPrinter(language.English)

// Number formats n with thousands separators, e.g. 12,345
func Number(n int) string {
	return englishPrinter.Sprintf("%d", n)
}
