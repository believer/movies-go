package tmdbtest

import (
	"believer/movies/services/tmdb"
	"believer/movies/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFake(t *testing.T) {
	var fake Fake
	fake.Add(types.MovieDetailsResponse{Title: "Inception"}, types.MovieCreditsResponse{}, "tt1375666", "27205")

	details, _, err := fake.Fetch("27205")
	assert.NoError(t, err)
	assert.Equal(t, "Inception", details.Title)

	_, _, err = fake.Fetch("tt0000000")
	assert.ErrorIs(t, err, tmdb.ErrNotFound)

	assert.Equal(t, []string{"27205", "tt0000000"}, fake.Calls())
}
