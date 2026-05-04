package db

import (
	"believer/movies/types"
	"believer/movies/utils"

	"github.com/jmoiron/sqlx"
)

type ProductionItem struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func (p ProductionItem) Title() string {
	return p.Name
}

func (p ProductionItem) Subtitle() string {
	return ""
}

func (p ProductionItem) Href() string {
	return utils.CreateSelfHealingUrl("production-country", p.Name, p.ID)
}

// Repo
// =====================================================

type ProductionCountryQuerier interface {
	ListProductionCountries() ([]ProductionItem, error)
	GetProductionCountryName(id string) (TableName, error)
	GetProductionCountryMovies(id, userID string, offset int) (types.Movies, error)
	GetProductionCountryStats(userID, year string) ([]types.ListItem, error)
}

type ProductionCountryRepository struct {
	db *sqlx.DB
}

func NewProductionCountryRepository(db *sqlx.DB) *ProductionCountryRepository {
	return &ProductionCountryRepository{db}
}

func (r *ProductionCountryRepository) ListProductionCountries() ([]ProductionItem, error) {
	var items []ProductionItem
	err := r.db.Select(&items, listProductionCountriesQuery)
	return items, err
}

func (r *ProductionCountryRepository) GetProductionCountryName(id string) (TableName, error) {
	var item TableName
	err := r.db.Get(&item, productionCountryNameQuery, id)
	return item, err
}

func (r *ProductionCountryRepository) GetProductionCountryMovies(id, userID string, offset int) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, productionCountryMoviesQuery, id, userID, offset)
	return movies, err
}

func (r *ProductionCountryRepository) GetProductionCountryStats(userID, year string) ([]types.ListItem, error) {
	var stats []types.ListItem
	err := r.db.Select(&stats, productionCountryStatsQuery, userID, year)
	return stats, err
}

// Queries
// =====================================================

const listProductionCountriesQuery = `
SELECT
    id,
    name
FROM
    production_country
ORDER BY
    name ASC
`

const productionCountryNameQuery = `
SELECT
    name
FROM
    production_country
WHERE
    id = $1
`

const productionCountryMoviesQuery = `
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie_country AS t
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
    t.country_id = $1
ORDER BY
    m.release_date DESC OFFSET $3
LIMIT 50
	`

const productionCountryStatsQuery = `
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
    INNER JOIN movie_country t ON t.movie_id = s.movie_id
    INNER JOIN production_country pc ON pc.id = t.country_id
GROUP BY
    pc.id
ORDER BY
    count DESC
LIMIT 10
	`
