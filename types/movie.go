package types

import (
	"believer/movies/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/a-h/templ"
)

type CastAndCrew struct {
	Name string `db:"name"`
	Job  string `db:"job"`
}

type MovieGenre struct {
	Name string `db:"name" json:"name"`
	ID   int    `db:"id" json:"id"`
}

func (g MovieGenre) LinkTo() string {
	return fmt.Sprintf("/genre/%d", g.ID)
}

type MovieGenres []MovieGenre

func (u *MovieGenres) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, u)
	case string:
		return json.Unmarshal([]byte(vv), u)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

type Movie struct {
	Cast           []CastAndCrew   `db:"cast"`
	CreatedAt      time.Time       `db:"created_at"`
	Genres         MovieGenres     `db:"genres"`
	ID             int             `db:"id" json:"id"`
	ImdbId         string          `db:"imdb_id"`
	ImdbRating     sql.NullFloat64 `db:"imdb_rating"`
	Overview       string          `db:"overview" json:"overview"`
	Poster         string          `db:"poster"`
	Rating         sql.NullInt64   `db:"rating"`
	ReleaseDate    time.Time       `db:"release_date" json:"release_date"`
	Runtime        int             `db:"runtime"`
	Series         sql.NullString  `db:"series"`
	NumberInSeries sql.NullInt64   `db:"number_in_series"`
	Tagline        string          `db:"tagline"`
	Title          string          `db:"title" json:"title"`
	UpdatedAt      time.Time       `db:"updated_at"`
	WatchedAt      time.Time       `db:"watched_at" json:"watchedAt"`
	Seen           bool            `db:"seen"`
}

// Format runtime in hours and minutes from minutes
func (m Movie) RuntimeFormatted() string {
	return utils.FormatRuntime(m.Runtime)
}

// The movie's release date formatted as ISO 8601 - YYYY-MM-DD
func (m Movie) ISOReleaseDate() string {
	return m.ReleaseDate.Format("2006-01-02")
}

// Link to the movie
func (m Movie) LinkTo() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/movie/%d", m.ID))
}

// Link to the movie's release year
func (m Movie) LinkToYear() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/year/%s", m.ReleaseDate.Format("2006")))
}

// Link to the movie's watchlist add
func (m Movie) LinkToWatchlistAdd() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/movie/new?imdbId=%s&id=%d", m.ImdbId, m.ID))
}

type Movies []Movie

func (u *Movies) Scan(v interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, u)
	case string:
		return json.Unmarshal([]byte(vv), u)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

func (m Movies) NumberOfMovies() string {
	numberOfMovies := len(m)
	text := "movies"

	if numberOfMovies == 1 {
		text = "movie"
	}

	return fmt.Sprintf("%d %s", numberOfMovies, text)
}

func (m Movies) NumberOfSeenMovies(seen int) string {
	numberOfMovies := m.NumberOfMovies()

	return fmt.Sprintf("Seen %d / %s", seen, numberOfMovies)
}
