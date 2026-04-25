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
    (s.id IS NOT NULL) AS "seen"
FROM
    rating r
    INNER JOIN movie m ON m.id = r.movie_id
    LEFT JOIN ( SELECT DISTINCT ON (movie_id)
            movie_id,
            id
        FROM
            public.seen
        WHERE
            user_id = $1
        ORDER BY
            movie_id,
            id) AS s ON m.id = s.movie_id
WHERE
    r.rating = $2
    AND r.user_id = $1
ORDER BY
    m.release_date DESC OFFSET $3
LIMIT 50
`
