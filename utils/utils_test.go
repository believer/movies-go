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
		{"TT0111161", "tt0111161", nil},
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

func TestFormatRuntime(t *testing.T) {
	tests := []struct {
		runtime int
		want    string
	}{
		{runtime: 0, want: ""},
		{runtime: 1, want: "1m"},
		{runtime: 2, want: "2m"},
		{runtime: 60, want: "1h"},
		{runtime: 61, want: "1h 1m"},
		{runtime: 120, want: "2h"},
		{runtime: 121, want: "2h 1m"},
		{runtime: 1440, want: "1d"},
		{runtime: 1441, want: "1d 1m"},
		{runtime: 1500, want: "1d 1h"},
		{runtime: 1501, want: "1d 1h 1m"},
	}

	for _, tc := range tests {
		got := FormatRuntime(tc.runtime)

		if got != tc.want {
			t.Errorf("FormatRuntime(%d) = %v; want %v", tc.runtime, got, tc.want)
		}
	}
}
