// Package tmdbtest provides an in-memory stand-in for the TMDB client.
package tmdbtest

import (
	"believer/movies/services/tmdb"
	"believer/movies/types"
	"sync"
)

type movie struct {
	details types.MovieDetailsResponse
	credits types.MovieCreditsResponse
}

// Fake serves movies added with Add and returns tmdb.ErrNotFound for
// everything else. The zero value is ready to use.
type Fake struct {
	mu     sync.Mutex
	movies map[string]movie
	calls  []string
}

// Add makes a movie available under each of the given IDs, typically its
// IMDb ID and its TMDB ID.
func (f *Fake) Add(details types.MovieDetailsResponse, credits types.MovieCreditsResponse, ids ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.movies == nil {
		f.movies = make(map[string]movie)
	}

	for _, id := range ids {
		f.movies[id] = movie{details, credits}
	}
}

func (f *Fake) Fetch(id string) (types.MovieDetailsResponse, types.MovieCreditsResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, id)

	m, ok := f.movies[id]

	if !ok {
		return types.MovieDetailsResponse{}, types.MovieCreditsResponse{}, tmdb.ErrNotFound
	}

	return m.details, m.credits, nil
}

// Calls returns the IDs Fetch was called with, in order.
func (f *Fake) Calls() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]string(nil), f.calls...)
}
