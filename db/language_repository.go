package db

import (
	"believer/movies/types"
	"believer/movies/utils"
	"strconv"

	"github.com/jmoiron/sqlx"
)

type Language struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	EnglishName string `db:"english_name"`
}

func (l Language) Title() string {
	return l.EnglishName
}

func (l Language) Subtitle() string {
	return l.Name
}

func (l Language) Href() string {
	return utils.CreateSelfHealingUrl("language", l.EnglishName, strconv.Itoa(l.ID))
}

// Repo
// =====================================================

type LanguageQuerier interface {
	ListLanguages() ([]Language, error)
	GetLanguageName(id string) (TableName, error)
	GetLanguageMovies(id, userID string, offset int) (types.Movies, error)
	GetLanguageStats(userID, year string) ([]types.ListItem, error)
}

type LanguageRepository struct {
	db *sqlx.DB
}

func NewLanguageRepository(db *sqlx.DB) *LanguageRepository {
	return &LanguageRepository{db}
}

func (r *LanguageRepository) ListLanguages() ([]Language, error) {
	var items []Language
	err := r.db.Select(&items, listLanguagesQuery)
	return items, err
}

func (r *LanguageRepository) GetLanguageName(id string) (TableName, error) {
	var item TableName
	err := r.db.Get(&item, languageNameQuery, id)
	return item, err
}

func (r *LanguageRepository) GetLanguageMovies(id, userID string, offset int) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, languageMoviesQuery, id, userID, offset)
	return movies, err
}

func (r *LanguageRepository) GetLanguageStats(userID, year string) ([]types.ListItem, error) {
	var stats []types.ListItem
	err := r.db.Select(&stats, languageStatsQuery, userID, year)
	return stats, err
}

// Queries
// =====================================================

const listLanguagesQuery = `
		SELECT
		    id,
		    name,
		    english_name
		FROM
		    "language"
		ORDER BY
		    english_name ASC
`

const languageNameQuery = `
SELECT
    name
FROM
    LANGUAGE
WHERE
    id = $1
`

const languageMoviesQuery = `
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie_language AS t
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
    t.language_id = $1
ORDER BY
    m.release_date DESC OFFSET $3
LIMIT 50
	`

const languageStatsQuery = `
SELECT
    pc.id,
    pc."name",
    pc.english_name AS link_name,
    COUNT(DISTINCT s.movie_id) AS count
FROM ( SELECT DISTINCT ON (movie_id)
        movie_id
    FROM
        seen
    WHERE
        user_id = $1
        AND ($2 = 'All'
            OR EXTRACT(YEAR FROM date) = $2::int)) AS s
    INNER JOIN movie_language t ON t.movie_id = s.movie_id
    INNER JOIN
    LANGUAGE pc
    ON pc.id = t.language_id
GROUP BY
    pc.id
ORDER BY
    count DESC
LIMIT 10
	`
