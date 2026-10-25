package tmdb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setupServer(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	original := baseURL
	baseURL = server.URL
	t.Cleanup(func() { baseURL = original })
}

func TestFetch(t *testing.T) {
	t.Run("returns details and credits", func(t *testing.T) {
		setupServer(t, func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/movie/tt1375666":
				w.Write([]byte(`{"id": 27205, "imdb_id": "tt1375666", "title": "Inception"}`))
			case "/movie/tt1375666/credits":
				w.Write([]byte(`{"cast": [{"id": 6193, "name": "Leonardo DiCaprio"}]}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})

		details, credits, err := Client{}.Fetch("tt1375666")

		assert.NoError(t, err)
		assert.Equal(t, "Inception", details.Title)
		assert.Equal(t, 27205, details.TmdbId)
		assert.Len(t, credits.Cast, 1)
		assert.Equal(t, "Leonardo DiCaprio", credits.Cast[0].Name)
	})

	t.Run("returns ErrNotFound for unknown movie", func(t *testing.T) {
		setupServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		_, _, err := Client{}.Fetch("tt0000000")

		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("returns error for unexpected status", func(t *testing.T) {
		setupServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		_, _, err := Client{}.Fetch("tt1375666")

		assert.Error(t, err)
		assert.NotErrorIs(t, err, ErrNotFound)
	})
}

func TestWatchProviders(t *testing.T) {
	t.Run("returns empty response when not found", func(t *testing.T) {
		setupServer(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		providers, err := New("tt1375666").WatchProviders()

		assert.NoError(t, err)
		assert.Empty(t, providers.Results.SE.Buy)
	})
}
