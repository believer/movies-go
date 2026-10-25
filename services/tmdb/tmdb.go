package tmdb

import (
	"believer/movies/types"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

var baseURL = "https://api.themoviedb.org/3"

// ErrNotFound is returned when TMDB responds with 404.
var ErrNotFound = errors.New("tmdb: not found")

// Client fetches movies from TMDB over HTTP.
type Client struct{}

// Fetch returns a movie's details and credits. The id is an IMDb or TMDB ID.
// It returns ErrNotFound when TMDB doesn't know the movie.
func (Client) Fetch(id string) (types.MovieDetailsResponse, types.MovieCreditsResponse, error) {
	t := New(id)

	details, err := t.Movie()

	if err != nil {
		return types.MovieDetailsResponse{}, types.MovieCreditsResponse{}, fmt.Errorf("fetch movie %s: %w", id, err)
	}

	credits, err := t.Credits()

	if err != nil {
		return types.MovieDetailsResponse{}, types.MovieCreditsResponse{}, fmt.Errorf("fetch credits %s: %w", id, err)
	}

	return details, credits, nil
}

type Tmdb struct {
	ID string
}

func New(id string) *Tmdb {
	return &Tmdb{
		ID: id,
	}
}

func (t *Tmdb) Movie() (types.MovieDetailsResponse, error) {
	return fetchJSON[types.MovieDetailsResponse](
		[]string{"movie", t.ID},
		map[string]string{},
	)
}

func (t *Tmdb) Credits() (types.MovieCreditsResponse, error) {
	return fetchJSON[types.MovieCreditsResponse](
		[]string{"movie", t.ID, "credits"},
		map[string]string{},
	)
}

func (t *Tmdb) Search(query string) (types.SearchMovieResponse, error) {
	return fetchJSON[types.SearchMovieResponse](
		[]string{"search", "movie"},
		map[string]string{"query": query},
	)
}

// WatchProviders returns an empty response when TMDB has no providers.
func (t *Tmdb) WatchProviders() (types.MovieWatchProvidersResponse, error) {
	providers, err := fetchJSON[types.MovieWatchProvidersResponse](
		[]string{"movie", t.ID, "watch", "providers"},
		map[string]string{},
	)

	if errors.Is(err, ErrNotFound) {
		return types.MovieWatchProvidersResponse{}, nil
	}

	return providers, err
}

func fetchJSON[T any](path []string, additionalParams map[string]string) (result T, err error) {
	// Construct URL
	tmdbKey := os.Getenv("TMDB_API_KEY")

	u, err := url.JoinPath(baseURL, path...)

	if err != nil {
		return result, err
	}

	params := url.Values{
		"api_key": {tmdbKey},
	}

	for k, v := range additionalParams {
		params.Add(k, v)
	}

	u = u + "?" + params.Encode()

	// Fetch data
	resp, err := httpClient.Get(u)

	if err != nil {
		return result, err
	}

	defer func() {
		cerr := resp.Body.Close()

		if err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return result, ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("TMDB API returned unexpected status code %d: %s", resp.StatusCode, resp.Status)
	}

	// Unmarshal JSON
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return result, err
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return result, err
	}

	return result, nil
}
