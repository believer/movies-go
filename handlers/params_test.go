package handlers

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseID(t *testing.T) {
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
			got, err := ParseID(tc.url)
			assert.Equal(t, tc.want, got)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAvailableYears(t *testing.T) {
	fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	got := AvailableYears(fixedTime)

	want := []string{"2024", "2023", "2022", "2021", "2020", "2019", "2018", "2017", "2016", "2015", "2014", "2013", "2012"}

	assert.Equal(t, want, got)
}

func TestParseWatchedAt(t *testing.T) {
	now := time.Date(2026, 10, 25, 21, 15, 30, 0, time.UTC)

	tests := []struct {
		value string
		want  time.Time
	}{
		{"2026-10-20T19:45", time.Date(2026, 10, 20, 19, 45, 0, 0, time.UTC)},
		{"2026-10-20", time.Date(2026, 10, 20, 21, 15, 30, 0, time.UTC)},
		{"", now},
		{"yesterday", now},
	}

	for _, tc := range tests {
		t.Run(tc.value, func(t *testing.T) {
			assert.Equal(t, tc.want, ParseWatchedAt(tc.value, now))
		})
	}
}
