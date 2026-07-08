package db

import (
	t "believer/movies/types"

	"github.com/jmoiron/sqlx"
)

// Repo
// =====================================================

type ListQuerier interface {
	GetList(id string) (t.List, error)
	GetLists() ([]t.List, error)
	GetListMovies(id, userID string) (t.Movies, error)
	GetListsByMovieID(movieID string) ([]t.List, error)
}

type ListRepository struct {
	db *sqlx.DB
}

func NewListRepository(db *sqlx.DB) *ListRepository {
	return &ListRepository{db}
}
func (r *ListRepository) GetLists() ([]t.List, error) {
	var lists []t.List
	err := r.db.Select(&lists, listsQuery)
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

const listsQuery = `SELECT id, name, source, slug FROM official_list ORDER BY name ASC`
const listQuery = `SELECT id, name, description, slug, source FROM official_list WHERE id = $1`

const listMoviesQuery = `
SELECT
	l.rank,
	m.id,
	m.title,
	(s.id IS NOT NULL) AS "seen"
FROM
	official_list_movie l
	INNER JOIN movie m ON m.id = l.movie_id
	LEFT JOIN (
		SELECT DISTINCT
			ON (movie_id) movie_id,
			id
		FROM
			public.seen
		WHERE
			user_id = $2
		ORDER BY
			movie_id,
			id
	) AS s ON m.id = s.movie_id
WHERE
	list_id = $1
ORDER BY
	RANK ASC
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
