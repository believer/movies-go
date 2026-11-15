package db

import (
	"believer/movies/types"

	"github.com/jmoiron/sqlx"
)

// Repo
// =====================================================

type RatingsQuerier interface {
	GetMoviesByRating(userID string, rating, page int) (types.Movies, error)
}

type RatingsRepository struct {
	db *sqlx.DB
}

func NewRatingsRepository(db *sqlx.DB) *RatingsRepository {
	return &RatingsRepository{db}
}

func (r *RatingsRepository) GetMoviesByRating(userID string, rating, page int) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, moviesByRatingQuery, userID, rating, page)
	return movies, err
}

// Queries
// =====================================================

const moviesByRatingQuery = `
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    m.imdb_id,
    (w.movie_id IS NOT NULL) AS "seen"
FROM
    latest_rating AS r
    INNER JOIN movie m ON m.id = r.movie_id
    LEFT JOIN watched_movie AS w ON w.movie_id = m.id
        AND w.user_id = $1
WHERE
    r.user_id = $1
    AND r.rating = $2
ORDER BY
    m.release_date DESC,
    m.id OFFSET $3
LIMIT 50
`
