package tmdb

import (
	"believer/movies/types"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

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

func (t *Tmdb) WatchProviders() (types.MovieWatchProvidersResponse, error) {
	return fetchJSON[types.MovieWatchProvidersResponse](
		[]string{"movie", t.ID, "watch", "providers"},
		map[string]string{},
	)
}

func fetchJSON[T any](path []string, additionalParams map[string]string) (T, error) {
	var result T

	// Construct URL
	tmdbKey := os.Getenv("TMDB_API_KEY")
	baseUrl := "https://api.themoviedb.org/3"

	u, err := url.JoinPath(baseUrl, path...)

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
	resp, err := http.Get(u)

	if err != nil {
		return result, err
	}

	defer func() {
		cerr := resp.Body.Close()

		if err != nil {
			err = cerr
		}
	}()

	if resp.StatusCode == 404 {
		log.Printf("Movie watch providers not found")
	}

	// Unmarshal JSON
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return result, err
	}

	if err := json.Unmarshal([]byte(body), &result); err != nil {
		return result, err
	}

	return result, nil
}
