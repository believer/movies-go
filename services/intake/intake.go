// Package intake brings movies into the diary from TMDB and keeps them up to
// date. See Movie intake in GLOSSARY.md.
package intake

import (
	"believer/movies/services/tmdb"
	"believer/movies/types"
	"believer/movies/utils/awards"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

var (
	// ErrMovieNotFound is returned when a movie is neither in the database
	// nor on TMDB.
	ErrMovieNotFound = errors.New("movie not found")

	// ErrAlreadyInSeries is returned when the movie already has the given
	// number in the series.
	ErrAlreadyInSeries = errors.New("movie already in series")
)

// MovieSource fetches a movie's details and credits by IMDb or TMDB ID, and
// returns tmdb.ErrNotFound for unknown movies. tmdb.Client is the production
// source, tmdbtest.Fake the one used in tests.
type MovieSource interface {
	Fetch(id string) (types.MovieDetailsResponse, types.MovieCreditsResponse, error)
}

type Intake struct {
	db     *sqlx.DB
	source MovieSource
}

func New(db *sqlx.DB, source MovieSource) *Intake {
	return &Intake{db, source}
}

// MovieRef identifies a movie, either by its ID in the database or by an
// IMDb or TMDB ID.
type MovieRef struct {
	id       int
	external string
}

// ByID refers to a movie already in the database.
func ByID(id int) MovieRef {
	return MovieRef{id: id}
}

// ByExternalID refers to a movie by IMDb ID (tt…) or TMDB ID (digits).
func ByExternalID(id string) MovieRef {
	return MovieRef{external: id}
}

// SeriesEntry places a movie in a series. ID refers to an existing series;
// when it's zero, the series named Name is used, and created if missing.
type SeriesEntry struct {
	ID     int
	Name   string
	Number float64
}

type MovieOptions struct {
	// WilhelmScream is only recorded when the movie is first added.
	WilhelmScream bool
	Series        *SeriesEntry
}

// EnsureMovie makes sure the movie is in the database and returns its ID. A
// movie that is missing is fetched from TMDB and added with its cast, crew,
// genres, languages, countries, production companies and awards. A movie
// that already exists is not refreshed; use RefreshMovie for that.
func (i *Intake) EnsureMovie(ref MovieRef, opts MovieOptions) (int, error) {
	p, err := i.resolve(ref)

	if err != nil {
		return 0, err
	}

	tx, err := i.db.Beginx()

	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	id, err := i.write(tx, p, opts)

	if err != nil {
		return 0, err
	}

	return id, tx.Commit()
}

// RefreshMovie fetches the movie from TMDB again and overwrites its details,
// cast, crew, genres, languages, countries, production companies and awards.
func (i *Intake) RefreshMovie(movieID int) error {
	var imdbID string

	err := i.db.Get(&imdbID, `SELECT imdb_id FROM movie WHERE id = $1`, movieID)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrMovieNotFound
	}

	if err != nil {
		return err
	}

	details, credits, err := i.fetch(imdbID)

	if err != nil {
		return err
	}

	tx, err := i.db.Beginx()

	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(`
UPDATE
    movie
SET
    title = $2,
    runtime = $3,
    release_date = NULLIF ($4, '')::date,
    imdb_id = $5,
    overview = $6,
    poster = $7,
    tagline = $8,
    updated_at = NOW(),
    tmdb_id = $9
WHERE
    id = $1`,
		movieID,
		details.Title,
		details.Runtime,
		details.ReleaseDate,
		details.ImdbId,
		details.Overview,
		details.Poster,
		details.Tagline,
		details.TmdbId,
	)

	if err != nil {
		return err
	}

	if err := writeMetadata(tx, movieID, details, credits); err != nil {
		return err
	}

	slog.Info("Movie refreshed", "id", movieID, "title", details.Title)

	return tx.Commit()
}

// pending is a movie resolved ahead of a write: either one already in the
// database, or one fetched from TMDB and not yet inserted. Fetching happens
// before the transaction opens, so no HTTP call holds a transaction.
type pending struct {
	id      int
	details types.MovieDetailsResponse
	credits types.MovieCreditsResponse
}

func (i *Intake) resolve(ref MovieRef) (pending, error) {
	var query string
	var arg any

	switch {
	case ref.id != 0:
		query, arg = `SELECT id FROM movie WHERE id = $1`, ref.id
	case strings.HasPrefix(ref.external, "tt"):
		query, arg = `SELECT id FROM movie WHERE imdb_id = $1`, ref.external
	case ref.external != "":
		query, arg = `SELECT id FROM movie WHERE tmdb_id = $1::int`, ref.external
	default:
		return pending{}, ErrMovieNotFound
	}

	var id int
	err := i.db.Get(&id, query, arg)

	if err == nil {
		return pending{id: id}, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return pending{}, err
	}

	if ref.id != 0 {
		return pending{}, ErrMovieNotFound
	}

	details, credits, err := i.fetch(ref.external)

	if err != nil {
		return pending{}, err
	}

	return pending{details: details, credits: credits}, nil
}

func (i *Intake) fetch(id string) (types.MovieDetailsResponse, types.MovieCreditsResponse, error) {
	details, credits, err := i.source.Fetch(id)

	if errors.Is(err, tmdb.ErrNotFound) {
		return details, credits, ErrMovieNotFound
	}

	return details, credits, err
}

// write inserts a fetched movie, or keeps an existing one, and applies the
// options. It returns the movie's ID.
func (i *Intake) write(tx *sqlx.Tx, p pending, opts MovieOptions) (int, error) {
	id := p.id

	if id == 0 {
		var err error
		id, err = insertMovie(tx, p.details, p.credits, opts.WilhelmScream)

		if err != nil {
			return 0, err
		}
	}

	if opts.Series != nil {
		if err := addToSeries(tx, id, *opts.Series); err != nil {
			return 0, err
		}
	}

	return id, nil
}

func insertMovie(tx *sqlx.Tx, details types.MovieDetailsResponse, credits types.MovieCreditsResponse, wilhelm bool) (int, error) {
	var id int

	// A movie looked up by TMDB ID can already exist under its IMDb ID
	err := tx.Get(
		&id,
		`
INSERT INTO movie (title, runtime, release_date, imdb_id, overview, poster, tagline, tmdb_id, wilhelm)
    VALUES ($1, $2, NULLIF ($3, '')::date, $4, $5, $6, $7, $8, $9)
ON CONFLICT (imdb_id)
    DO UPDATE SET
        title = excluded.title,
        runtime = excluded.runtime,
        release_date = excluded.release_date,
        imdb_id = excluded.imdb_id,
        overview = excluded.overview,
        poster = excluded.poster,
        tagline = excluded.tagline,
        tmdb_id = excluded.tmdb_id
    RETURNING
        id
		`,
		details.Title,
		details.Runtime,
		details.ReleaseDate,
		details.ImdbId,
		details.Overview,
		details.Poster,
		details.Tagline,
		details.TmdbId,
		wilhelm,
	)

	if err != nil {
		return 0, fmt.Errorf("insert movie: %w", err)
	}

	if err := writeMetadata(tx, id, details, credits); err != nil {
		return 0, err
	}

	slog.Info("Movie added", "id", id, "title", details.Title)

	return id, nil
}

// writeMetadata writes everything that comes with a movie from TMDB. Awards
// go last since they are matched against the movie's people.
func writeMetadata(tx *sqlx.Tx, id int, details types.MovieDetailsResponse, credits types.MovieCreditsResponse) error {
	if err := addLanguages(tx, id, details); err != nil {
		return err
	}

	if err := addGenres(tx, id, details); err != nil {
		return err
	}

	if err := addCountries(tx, id, details); err != nil {
		return err
	}

	if err := addProductionCompanies(tx, id, details); err != nil {
		return err
	}

	if err := addCast(tx, credits, id); err != nil {
		return err
	}

	awards.AddOscars(tx, details.ImdbId)
	awards.AddBaftas(tx, details.ImdbId)

	return nil
}

func addToSeries(tx *sqlx.Tx, movieID int, entry SeriesEntry) error {
	seriesID := entry.ID

	if seriesID == 0 {
		err := tx.Get(&seriesID, `
INSERT INTO series (name)
    VALUES ($1)
ON CONFLICT (name)
    DO UPDATE SET
        name = EXCLUDED.name
    RETURNING
        id`, entry.Name)

		if err != nil {
			return fmt.Errorf("insert series: %w", err)
		}
	}

	_, err := tx.Exec(`
INSERT INTO movie_series (movie_id, series_id, number_in_series)
    VALUES ($1, $2, $3)`, movieID, seriesID, entry.Number)

	var pqErr *pq.Error

	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrAlreadyInSeries
	}

	return err
}
