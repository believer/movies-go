package db

import (
	"believer/movies/types"
	"believer/movies/utils"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type Genre struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func (g Genre) Title() string {
	return g.Name
}

func (g Genre) Subtitle() string {
	return ""
}

func (g Genre) Href() string {
	return utils.CreateSelfHealingUrl("genre", g.Name, strconv.Itoa(g.ID))
}

// Repo
// =====================================================

type GenreQuerier interface {
	ListGenres() ([]Genre, error)
	GetGenreName(id string) (TableName, error)
	GetGenreMovies(id, userID string, offset int) (types.Movies, error)
	GetGenreStats(userID, year string) ([]types.ListItem, error)
}

type GenreRepository struct {
	db *sqlx.DB
}

func NewGenreRepository(db *sqlx.DB) *GenreRepository {
	return &GenreRepository{db}
}

func (r *GenreRepository) ListGenres() ([]Genre, error) {
	var items []Genre
	err := r.db.Select(&items, listGenreQuery)
	return items, err
}

func (r *GenreRepository) GetGenreName(id string) (TableName, error) {
	var item TableName
	err := r.db.Get(&item, genreNameQuery, id)
	return item, err
}

func (r *GenreRepository) GetGenreMovies(id, userID string, offset int) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, genreMoviesQuery, id, userID, offset)
	return movies, err
}

func (r *GenreRepository) GetGenreStats(userID, year string) ([]types.ListItem, error) {
	var stats []types.ListItem
	err := r.db.Select(&stats, genreStatsQuery, userID, year)
	return stats, err
}

// Queries
// =====================================================

const listGenreQuery = `
		SELECT
		    id,
		    name
		FROM
		    genre
		ORDER BY
		    name ASC
`

const genreNameQuery = `
SELECT
    name
FROM
    genre
WHERE
    id = $1
`

const genreMoviesQuery = `
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie_genre AS t
    INNER JOIN movie m ON m.id = t.movie_id
    LEFT JOIN ( SELECT DISTINCT ON (movie_id)
            movie_id,
            id
        FROM
            public.seen
        WHERE
            user_id = $2
        ORDER BY
            movie_id,
            id) AS s ON m.id = s.movie_id
WHERE
    t.genre_id = $1
ORDER BY
    m.release_date DESC OFFSET $3
LIMIT 50
	`

const genreStatsQuery = `
SELECT
    pc.id,
    pc."name",
    COUNT(DISTINCT s.movie_id) AS count
FROM ( SELECT DISTINCT ON (movie_id)
        movie_id
    FROM
        seen
    WHERE
        user_id = $1
        AND ($2 = 'All'
            OR EXTRACT(YEAR FROM date) = $2::int)) AS s
    INNER JOIN movie_genre t ON t.movie_id = s.movie_id
    INNER JOIN genre pc ON pc.id = t.genre_id
GROUP BY
    pc.id
ORDER BY
    count DESC
LIMIT 10
	`
