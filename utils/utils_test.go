package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
		{"https://www.imdb.com/", "", fmt.Errorf("empty ID")},
		{"TT0111161", "tt0111161", nil},
		{"321405", "321405", nil},
		{"", "", fmt.Errorf("empty ID")},
		{"not_a_url", "", fmt.Errorf("invalid ID format: not_a_url")},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			got, err := ParseId(tc.url)
			assert.Equal(t, tc.want, got)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFormatRuntime(t *testing.T) {
	tests := []struct {
		runtime int
		want    string
	}{
		{0, "0m"},
		{1, "1m"},
		{2, "2m"},
		{60, "1h"},
		{61, "1h 1m"},
		{120, "2h"},
		{121, "2h 1m"},
		{1440, "1d"},
		{1441, "1d 1m"},
		{1500, "1d 1h"},
		{1501, "1d 1h 1m"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, FormatRuntime(tc.runtime))
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		text string
		want string
	}{
		{"Hugh Jackman", "hugh-jackman"},
		{"Alfonso Cuarón", "alfonso-cuaron"},
	}
	for _, tc := range tests {
		t.Run(tc.text, func(t *testing.T) {
			assert.Equal(t, tc.want, Slugify(tc.text))
		})
	}
}
