package types

import (
	"believer/movies/utils"
	"database/sql"
	"fmt"
	"time"
)

type Entity struct {
	Name string `db:"name" json:"name"`
	ID   int    `db:"id" json:"id"`
}

func (e Entity) LinkTo(prefix string) string {
	return fmt.Sprintf("/%s/%s-%d", prefix, utils.Slugify(e.Name), e.ID)
}

// Movie sub structs
// ======================================================

type CastAndCrew struct {
	Name string `db:"name" json:"name"`
	Job  string `db:"job" json:"job"`
}

// Genre
type MovieGenre struct{ Entity }
type MovieGenres []MovieGenre

func (u *MovieGenres) Scan(v any) error {
	return utils.ScanJSON(v, u)
}

// Language
type MovieLanguage struct{ Entity }
type MovieLanguages []MovieLanguage

func (u *MovieLanguages) Scan(v any) error {
	return utils.ScanJSON(v, u)
}

// Production Companies
type MovieProductionCompany struct{ Entity }

func (mpc *MovieProductionCompany) LinkTo() string {
	return fmt.Sprintf("/production-company/%s-%d", utils.Slugify(mpc.Name), mpc.ID)
}

type MovieProductionCompanies []MovieProductionCompany

func (u *MovieProductionCompanies) Scan(v any) error {
	return utils.ScanJSON(v, u)
}

// Production Countries
type MovieProductionCountry struct {
	Id   string `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (mpc *MovieProductionCountry) LinkTo() string {
	return fmt.Sprintf("/production-country/%s-%s", utils.Slugify(mpc.Name), mpc.Id)
}

type MovieProductionCountries []MovieProductionCountry

func (u *MovieProductionCountries) Scan(v any) error {
	return utils.ScanJSON(v, u)
}

// Movie
// ======================================================

type Movie struct {
	Cast                []CastAndCrew            `db:"cast" json:"cast"`
	CreatedAt           time.Time                `db:"created_at" json:"createdAt"`
	Genres              MovieGenres              `db:"genres" json:"genres"`
	ID                  int                      `db:"id" json:"id"`
	ImdbId              string                   `db:"imdb_id" json:"imdbId"`
	ImdbRating          sql.NullFloat64          `db:"imdb_rating" json:"imdbRating"`
	Languages           MovieLanguages           `db:"languages" json:"languages"`
	NumberInSeries      utils.NullInt64          `db:"number_in_series" json:"numberInSeries"`
	OriginalTitle       sql.NullString           `db:"original_title" json:"originaTitle"`
	Overview            string                   `db:"overview" json:"overview"`
	Poster              string                   `db:"poster" json:"poster"`
	ProductionCompanies MovieProductionCompanies `db:"production_companies"`
	ProductionCountries MovieProductionCountries `db:"production_countries"`
	RatedAt             sql.NullTime             `db:"rated_at" json:"ratedAt"`
	Rating              sql.NullInt64            `db:"rating" json:"rating"`
	ReleaseDate         utils.NullTime           `db:"release_date" json:"releaseDate"`
	Runtime             int                      `db:"runtime" json:"runtime"`
	Seen                bool                     `db:"seen" json:"seen"`
	Series              sql.NullString           `db:"series" json:"series"`
	SeriesID            sql.NullInt64            `db:"series_id" json:"seriesId"`
	Tagline             string                   `db:"tagline" json:"tagline"`
	Title               string                   `db:"title" json:"title"`
	UpdatedAt           time.Time                `db:"updated_at" json:"updatedAt"`
	WatchedAt           sql.NullTime             `db:"watched_at" json:"watchedAt"`
	WilhelmScream       sql.NullBool             `db:"wilhelm" json:"wilhelm"`
}

// Format runtime in hours and minutes from minutes
func (m Movie) RuntimeFormatted() string {
	return utils.FormatRuntime(m.Runtime)
}

// The movie's release date formatted as ISO 8601 - YYYY-MM-DD
func (m Movie) ISOReleaseDate() string {
	if !m.ReleaseDate.Valid {
		return ""
	}

	return m.ReleaseDate.Time.Format("2006-01-02")
}

func (m Movie) ISOCreatedDate() string {
	return m.CreatedAt.Format("2006-01-02")
}

func (m Movie) ISOWatchedAt() string {
	if !m.WatchedAt.Valid {
		return ""
	}

	return m.WatchedAt.Time.Format("2006-01-02")
}

func (m Movie) ReleaseDateOrCreatedAt() string {
	if !m.ReleaseDate.Valid {
		return ""
	}

	if m.ReleaseDate.Time.Year() == 1 {
		return m.CreatedAt.Format("2006-01-02")
	}

	return m.ISOReleaseDate()
}

// Release year
func (m Movie) ReleaseYear() string {
	if !m.ReleaseDate.Valid {
		return ""
	}

	return m.ReleaseDate.Time.Format("2006")
}

// Link to the movie
func (m Movie) LinkTo() string {
	return fmt.Sprintf("/movie/%s-%d", utils.Slugify(m.Title), m.ID)
}

func (m Movie) LinkToReleaseYear() string {
	return fmt.Sprintf("/year/%s", m.ReleaseDate.Time.Format("2006"))
}

func (m Movie) LinkToCreatedYear() string {
	return fmt.Sprintf("/year/%s", m.CreatedAt.Format("2006"))
}

// Link to the movie's release year
func (m Movie) LinkToYear() string {
	year := m.ReleaseDate.Time

	if year.Year() == 1 {
		year = m.CreatedAt
	}

	return fmt.Sprintf("/year/%s", year.Format("2006"))
}

// Link to the movie's watchlist add
func (m Movie) LinkToWatchlistAdd() string {
	return fmt.Sprintf("/movie/new?imdbId=%s&id=%d", m.ImdbId, m.ID)
}

// Link to the movie's series
func (m Movie) LinkToSeries() string {
	if m.SeriesID.Valid && m.Series.Valid {
		return fmt.Sprintf("/series/%s-%d", utils.Slugify(m.Series.String), m.SeriesID.Int64)
	}

	return ""
}

// Movies
// ======================================================

type Movies []Movie

func (u *Movies) Scan(v any) error {
	return utils.ScanJSON(v, u)
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
