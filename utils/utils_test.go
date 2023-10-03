package utils

import (
	"fmt"
	"testing"
)

func TestParseImdbId(t *testing.T) {
	tests := []struct {
		url  string
		want string
		err  error
	}{
		{"https://www.imdb.com/title/tt0111161/", "tt0111161", nil},
		{"https://www.imdb.com/title/tt0068646", "tt0068646", nil},
		{"https://www.imdb.com/title/tt0339230/?ref_=ext_shr_lnk", "tt0339230", nil},
		{"https://www.imdb.com/", "", fmt.Errorf("Empty IMDb ID")},
		{"", "", fmt.Errorf("Empty IMDb ID")},
		{"not_a_url", "", fmt.Errorf("Invalid IMDb ID format: not_a_url")},
	}

	for _, tc := range tests {
		got, err := ParseImdbId(tc.url)

		if got != tc.want {
			t.Errorf("ParseImdbId(%q) = %v; want %v", tc.url, got, tc.want)
		}

		if (err != nil && tc.err == nil) || (err == nil && tc.err != nil) || (err != nil && err.Error() != tc.err.Error()) {
			t.Errorf("Expected error: %v, got: %v", tc.err, err)
		}
	}
}
