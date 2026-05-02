package db

import (
	"believer/movies/types"

	"github.com/jmoiron/sqlx"
)

// Repo
// =====================================================

type YearsQuerier interface {
	GetMoviesByYear(userID, year string, page int) (types.Movies, error)
}

type YearsRepository struct {
	db *sqlx.DB
}

func NewYearsRepository(db *sqlx.DB) *YearsRepository {
	return &YearsRepository{db}
}

func (r *YearsRepository) GetMoviesByYear(userID, year string, page int) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, moviesByYearQuery, userID, year, (page-1)*50)
	return movies, err
}

// Queries
// =====================================================

const moviesByYearQuery = `
SELECT
    m.id,
    m.title,
    m.release_date,
    m.imdb_id,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie AS m
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
    date_part('year', release_date) = $2
ORDER BY
    release_date ASC OFFSET $3
LIMIT 50
`
