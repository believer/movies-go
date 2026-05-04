package db

import (
	"believer/movies/types"

	"github.com/jmoiron/sqlx"
)

// Repo
// =====================================================

type WatchlistQuerier interface {
	GetUnreleasedMovies(userID, sortOrder string) (types.Movies, error)
	GetReleasedMovies(userID, sortOrder string) (types.Movies, error)
	GetTBDMovies(userID string) (types.Movies, error)
	DeleteFromWatchlist(id, userID string) error
}

type WatchlistRepository struct {
	db *sqlx.DB
}

func NewWatchlistRepository(db *sqlx.DB) *WatchlistRepository {
	return &WatchlistRepository{db}
}

func (r *WatchlistRepository) GetUnreleasedMovies(userID, sortOrder string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, getUnreleasedQuery, userID, sortOrder)
	return movies, err
}

func (r *WatchlistRepository) GetReleasedMovies(userID, sortOrder string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, getReleasedQuery, userID, sortOrder)
	return movies, err
}

func (r *WatchlistRepository) GetTBDMovies(userID string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, getTbdQuery, userID)
	return movies, err
}

func (r *WatchlistRepository) DeleteFromWatchlist(id, userID string) error {
	if _, err := r.db.Exec(deleteFromWatchlistQuery); err != nil {
		return err
	}

	return nil
}

// Queries
// =====================================================

const getUnreleasedQuery = `
SELECT
    m.id,
    m.title,
    m.imdb_id,
    m.release_date,
    w.created_at
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date > CURRENT_DATE
ORDER BY
    CASE WHEN $2 = 'Date added' THEN
        w.created_at
    ELSE
        m.release_date
    END ASC
`

const getReleasedQuery = `
SELECT
    m.id,
    m.title,
    m.imdb_id,
    m.release_date,
    w.created_at
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date <= CURRENT_DATE
ORDER BY
    CASE WHEN $2 = 'Release date' THEN
        m.release_date
    ELSE
        w.created_at
    END ASC
`

const getTbdQuery = `
SELECT
    m.id,
    m.title,
    m.imdb_id,
    m.release_date,
    w.created_at
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date IS NULL
`

const deleteFromWatchlistQuery = `
DELETE FROM watchlist
WHERE movie_id = $1
    AND user_id = $2
`
