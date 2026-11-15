package db

import (
	t "believer/movies/types"

	"github.com/jmoiron/sqlx"
)

// Repo
// =====================================================

type ListQuerier interface {
	GetList(id string) (t.List, error)
	GetLists(userID string) ([]t.List, error)
	GetListMovies(id, userID string) (t.Movies, error)
	GetListsByMovieID(movieID string) ([]t.List, error)
}

type ListRepository struct {
	db *sqlx.DB
}

func NewListRepository(db *sqlx.DB) *ListRepository {
	return &ListRepository{db}
}
func (r *ListRepository) GetLists(userID string) ([]t.List, error) {
	var lists []t.List
	err := r.db.Select(&lists, listsQuery, userID)
	return lists, err
}

func (r *ListRepository) GetList(id string) (t.List, error) {
	var list t.List
	err := r.db.Get(&list, listQuery, id)
	return list, err
}

func (r *ListRepository) GetListMovies(id, userID string) (t.Movies, error) {
	var movies t.Movies
	err := r.db.Select(&movies, listMoviesQuery, id, userID)
	return movies, err
}

func (r *ListRepository) GetListsByMovieID(movieID string) ([]t.List, error) {
	var lists []t.List
	err := r.db.Select(&lists, listsByMovieQuery, movieID)
	return lists, err
}

// Queries
// =====================================================

const listsQuery = `
SELECT
    ol.id,
    ol.name,
    ol.source,
    ol.slug,
    COUNT(*) AS total_movies,
    COUNT(w.movie_id) AS seen_movies,
    ROUND(100.0 * COUNT(w.movie_id) / COUNT(*), 2) AS percent_seen
FROM
    official_list ol
    INNER JOIN official_list_movie l ON l.list_id = ol.id
    INNER JOIN movie m ON m.id = l.movie_id
    LEFT JOIN watched_movie AS w ON w.movie_id = m.id
        AND w.user_id = $1
GROUP BY
    ol.id,
    ol.name
ORDER BY
    percent_seen DESC,
    ol.id`

const listQuery = `SELECT
    id,
    name,
    description,
    slug,
    source
FROM
    official_list
WHERE
    id = $1`

const listMoviesQuery = `
SELECT
    l.rank,
    m.id,
    m.title,
    (w.movie_id IS NOT NULL) AS "seen"
FROM
    official_list_movie l
    INNER JOIN movie m ON m.id = l.movie_id
    LEFT JOIN watched_movie AS w ON w.movie_id = m.id
        AND w.user_id = $2
WHERE
    list_id = $1
ORDER BY
    RANK ASC,
    m.id
`

const listsByMovieQuery = `
SELECT
    official_list.id,
    official_list.name,
    official_list.slug,
    official_list.source,
    official_list_movie.rank
FROM
    official_list_movie
    INNER JOIN official_list ON official_list_movie.list_id = official_list.id
WHERE
    official_list_movie.movie_id = $1
`
