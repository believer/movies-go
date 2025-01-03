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

type Entity struct {
	Name string `db:"name" json:"name"`
	ID   int    `db:"id" json:"id"`
}

func (e Entity) LinkTo(prefix string) string {
	return fmt.Sprintf("/%s/%s-%d", prefix, utils.Slugify(e.Name), e.ID)
}

// MovieGenre and MovieLanguage now embed Entity
type MovieGenre struct {
	Entity
}

type MovieLanguage struct {
	Entity
}

// MovieGenres and MovieLanguages share a common Scan implementation
type Scannable interface {
	Scan(v interface{}) error
}

func ScanEntity(v interface{}, dest interface{}) error {
	switch vv := v.(type) {
	case []byte:
		return json.Unmarshal(vv, dest)
	case string:
		return json.Unmarshal([]byte(vv), dest)
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}

type MovieGenres []MovieGenre

func (u *MovieGenres) Scan(v interface{}) error {
	return ScanEntity(v, u)
}

type MovieLanguages []MovieLanguage

func (u *MovieLanguages) Scan(v interface{}) error {
	return ScanEntity(v, u)
}

type Movie struct {
	Cast           []CastAndCrew   `db:"cast"`
	CreatedAt      time.Time       `db:"created_at"`
	Genres         MovieGenres     `db:"genres"`
	Languages      MovieLanguages  `db:"languages"`
	ID             int             `db:"id" json:"id"`
	ImdbId         string          `db:"imdb_id"`
	ImdbRating     sql.NullFloat64 `db:"imdb_rating"`
	NumberInSeries JSONNullInt64   `db:"number_in_series" json:"number_in_series"`
	OriginalTitle  sql.NullString  `db:"original_title"`
	Overview       string          `db:"overview" json:"overview"`
	Poster         string          `db:"poster"`
	Rating         sql.NullInt64   `db:"rating"`
	RatedAt        sql.NullTime    `db:"rated_at"`
	ReleaseDate    time.Time       `db:"release_date" json:"release_date"`
	Runtime        int             `db:"runtime"`
	Seen           bool            `db:"seen"`
	Series         sql.NullString  `db:"series" json:"series"`
	SeriesID       sql.NullInt64   `db:"series_id"`
	Tagline        string          `db:"tagline"`
	Title          string          `db:"title" json:"title"`
	UpdatedAt      time.Time       `db:"updated_at"`
	WatchedAt      time.Time       `db:"watched_at" json:"watchedAt"`
	WilhelmScream  sql.NullBool    `db:"wilhelm"`
}

// Format runtime in hours and minutes from minutes
func (m Movie) RuntimeFormatted() string {
	return utils.FormatRuntime(m.Runtime)
}

// The movie's release date formatted as ISO 8601 - YYYY-MM-DD
func (m Movie) ISOReleaseDate() string {
	return m.ReleaseDate.Format("2006-01-02")
}

func (m Movie) ReleaseDateOrCreatedAt() string {
	if m.ReleaseDate.Year() == 1 {
		return m.CreatedAt.Format("2006-01-02")
	}

	return m.ISOReleaseDate()
}

// Release year
func (m Movie) ReleaseYear() string {
	return m.ReleaseDate.Format("2006")
}

// Link to the movie
func (m Movie) LinkTo() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/movie/%s-%d", utils.Slugify(m.Title), m.ID))
}

// Link to the movie's release year
func (m Movie) LinkToYear() templ.SafeURL {
	year := m.ReleaseDate

	if year.Year() == 1 {
		year = m.CreatedAt
	}

	return templ.URL(fmt.Sprintf("/year/%s", year.Format("2006")))
}

// Link to the movie's watchlist add
func (m Movie) LinkToWatchlistAdd() templ.SafeURL {
	return templ.URL(fmt.Sprintf("/movie/new?imdbId=%s&id=%d", m.ImdbId, m.ID))
}

// Link to the movie's series
func (m Movie) LinkToSeries() templ.SafeURL {
	if m.SeriesID.Valid && m.Series.Valid {
		return templ.URL(fmt.Sprintf("/series/%s-%d", utils.Slugify(m.Series.String), m.SeriesID.Int64))
	}

	return ""
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

type JSONNullInt64 struct {
	sql.NullInt64
}

func (n *JSONNullInt64) UnmarshalJSON(data []byte) error {
	// Handle null JSON values
	if string(data) == "null" {
		n.Valid = false
		return nil
	}

	// Try to unmarshal an integer
	var intValue int64
	if err := json.Unmarshal(data, &intValue); err != nil {
		return err
	}

	// If successfully unmarshaled, assign the value
	n.Int64 = intValue
	n.Valid = true
	return nil
}

func (n JSONNullInt64) MarshalJSON() ([]byte, error) {
	// Return "null" if the value is not valid
	if !n.Valid {
		return json.Marshal(nil)
	}
	// Otherwise, return the int64 value
	return json.Marshal(n.Int64)
}
