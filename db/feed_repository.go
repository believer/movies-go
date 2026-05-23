package db

import (
	"believer/movies/types"

	"github.com/jmoiron/sqlx"
)

type FeedQuerier interface {
	SearchMovies(query string) (types.Movies, error)
	SearchPersons(query string, job string) (types.Persons, error)
	SearchMoviesByRating(rating string, userID string) (types.Movies, error)
	GetFeedMovies(userID string, offset int) (types.Movies, error)
}

type FeedRepository struct {
	db *sqlx.DB
}

func NewFeedRepository(db *sqlx.DB) *FeedRepository {
	return &FeedRepository{db}
}

func (r *FeedRepository) SearchMovies(query string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.overview,
    se.name AS "series",
    ms.number_in_series,
    m.release_date AS watched_at
FROM
    movie AS m
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
WHERE
    m.title ILIKE '%' || $1 || '%'
    OR m.original_title ILIKE '%' || $1 || '%'
    OR se.name ILIKE '%' || $1 || '%'
ORDER BY
    m.release_date DESC
	`, query)
	return movies, err
}

func (r *FeedRepository) SearchPersons(query string, job string) (types.Persons, error) {
	var persons types.Persons
	err := r.db.Select(&persons, `
SELECT
    p.id,
    p.name,
    count(*)
FROM
    person p
    INNER JOIN movie_person mp ON mp.person_id = p.id
WHERE
    p."name" ILIKE '%' || $1 || '%'
    AND mp.job = $2
GROUP BY
    p.id
ORDER BY
    COUNT DESC
LIMIT 100
	`, query, job)
	return persons, err
}

func (r *FeedRepository) SearchMoviesByRating(rating string, userID string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.overview,
    se.name AS "series",
    ms.number_in_series,
    m.release_date AS watched_at
FROM
    movie AS m
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
    LEFT JOIN rating AS r ON r.movie_id = m.id
WHERE
    r.rating = $1
    AND r.user_id = $2
ORDER BY
    m.release_date DESC
	`, rating, userID)
	return movies, err
}

func (r *FeedRepository) GetFeedMovies(userID string, offset int) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.overview,
    m.release_date,
    se.name AS "series",
    ms.number_in_series,
    s.date at time zone 'UTC' at time zone 'Europe/Stockholm' AS watched_at
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
WHERE
    user_id = $2
ORDER BY
    s.date DESC OFFSET $1
LIMIT 20
	`, offset, userID)
	return movies, err
}
