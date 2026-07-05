package db

import (
	"believer/movies/types"
	"believer/movies/utils"

	"github.com/jmoiron/sqlx"
)

type List struct {
	Description string `db:"description"`
	ID          string `db:"id"`
	Name        string `db:"name"`
	Slug        string `db:"slug"`
	Source      string `db:"source"`
}

func (l List) Title() string {
	return l.Name
}

func (l List) Subtitle() string {
	return l.Source
}

func (l List) Href() string {
	return utils.CreateSelfHealingUrl("list", l.Slug, l.ID)
}

// Repo
// =====================================================

type ListQuerier interface {
	GetList(id string) (List, error)
	GetLists() ([]List, error)
	GetListMovies(id, userID string) (types.Movies, error)
}

type ListRepository struct {
	db *sqlx.DB
}

func NewListRepository(db *sqlx.DB) *ListRepository {
	return &ListRepository{db}
}
func (r *ListRepository) GetLists() ([]List, error) {
	var lists []List
	err := r.db.Select(&lists, listsQuery)
	return lists, err
}

func (r *ListRepository) GetList(id string) (List, error) {
	var list List
	err := r.db.Get(&list, listQuery, id)
	return list, err
}

func (r *ListRepository) GetListMovies(id, userID string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, listMoviesQuery, id, userID)
	return movies, err
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
