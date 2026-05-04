package db

import (
	"believer/movies/types"

	"github.com/jmoiron/sqlx"
)

// Repo
// =====================================================

type ProductionCompanyQuerier interface {
	ListProductionCompanies(page int) ([]ProductionItem, error)
	GetProductionCompanyName(id string) (TableName, error)
	GetProductionCompanyMovies(id, userID string, offset int) (types.Movies, error)
	GetProductionCompanyStats(userID, year string) ([]types.ListItem, error)
}

type ProductionCompanyRepository struct {
	db *sqlx.DB
}

func NewProductionCompanyRepository(db *sqlx.DB) *ProductionCompanyRepository {
	return &ProductionCompanyRepository{db}
}

func (r *ProductionCompanyRepository) ListProductionCompanies(page int) ([]ProductionItem, error) {
	var items []ProductionItem
	err := r.db.Select(&items, listProductionCompaniesQuery, (page-1)*50)
	return items, err
}

func (r *ProductionCompanyRepository) GetProductionCompanyName(id string) (TableName, error) {
	var item TableName
	err := r.db.Get(&item, productionCompanyNameQuery, id)
	return item, err
}

func (r *ProductionCompanyRepository) GetProductionCompanyMovies(id, userID string, offset int) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, productionCompanyMoviesQuery, id, userID, offset)
	return movies, err
}

func (r *ProductionCompanyRepository) GetProductionCompanyStats(userID, year string) ([]types.ListItem, error) {
	var stats []types.ListItem
	err := r.db.Select(&stats, productionCompanyStatsQuery, userID, year)
	return stats, err
}

// Queries
// =====================================================

const listProductionCompaniesQuery = `
		SELECT
		    id,
		    name
		FROM
		    production_company
		ORDER BY
		    name ASC OFFSET $1
		LIMIT 50
`

const productionCompanyNameQuery = `
SELECT
    name
FROM
    production_company
WHERE
    id = $1
`

const productionCompanyMoviesQuery = `
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie_company AS t
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
    t.company_id = $1
ORDER BY
    m.release_date DESC OFFSET $3
LIMIT 50
	`

const productionCompanyStatsQuery = `
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
    INNER JOIN movie_company t ON t.movie_id = s.movie_id
    INNER JOIN production_company pc ON pc.id = t.company_id
GROUP BY
    pc.id
ORDER BY
    count DESC
LIMIT 10
	`
