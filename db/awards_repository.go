package db

import (
	"believer/movies/types"

	"github.com/jmoiron/sqlx"
)

// Repo
// =====================================================

type AwardsQuerier interface {
	GetByNominations(userID string, count int, awardType string) (types.Movies, error)
	GetByWins(userID string, count int, awardType string) (types.Movies, error)
	GetGroupedByMovie(year, awardType string) ([]types.AwardsByYear, error)
	GetGroupedByCategory(year, awardType string) ([]types.AwardsByCategory, error)
}

type AwardsRepository struct {
	db *sqlx.DB
}

func NewAwardsRepository(db *sqlx.DB) *AwardsRepository {
	return &AwardsRepository{db: db}
}

func (r *AwardsRepository) GetByNominations(userID string, count int, awardType string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, awardsByNominationsQuery, userID, count, awardType)
	return movies, err
}

func (r *AwardsRepository) GetByWins(userID string, count int, awardType string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, awardsByWinsQuery, userID, count, awardType)
	return movies, err
}

func (r *AwardsRepository) GetGroupedByMovie(year, awardType string) ([]types.AwardsByYear, error) {
	var awards []types.AwardsByYear
	err := r.db.Select(&awards, awardsGroupedByMovieQuery, year, awardType)
	return awards, err
}

func (r *AwardsRepository) GetGroupedByCategory(year, awardType string) ([]types.AwardsByCategory, error) {
	var awards []types.AwardsByCategory
	err := r.db.Select(&awards, awardsGroupedByCategoryQuery, year, awardType)
	return awards, err
}

// Queries
// =====================================================

const awardsByNominationsQuery = `
SELECT
    m.id,
    m.title,
    m.release_date,
    (s.id IS NOT NULL) AS "seen"
FROM
    award a
    INNER JOIN movie m ON m.imdb_id = a.imdb_id
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
    a.type = $3
GROUP BY
    a.imdb_id,
    m.id,
    s.id
HAVING
    count(DISTINCT CASE WHEN a.name IN ('Best Film', 'Best Screenplay', 'Editing', 'Adapted Screenplay') THEN
            a.name
        ELSE
            a.id::text
        END) = $2
ORDER BY
    m.release_date DESC`

const awardsByWinsQuery = `
SELECT
    m.id,
    m.title,
    m.release_date,
    (s.id IS NOT NULL) AS "seen"
FROM
    award a
    INNER JOIN movie m ON m.imdb_id = a.imdb_id
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
    winner = TRUE
    AND type = $3
GROUP BY
    a.imdb_id,
    m.id,
    s.id
HAVING
    count(DISTINCT a.name) = $2
ORDER BY
    m.release_date DESC
`

const awardsGroupedByMovieQuery = `
WITH nominees AS (
    SELECT
        a.imdb_id,
        a.name AS category,
        a.detail,
        a.winner,
        JSONB_AGG(
            CASE WHEN person IS NOT NULL
                AND person_id IS NOT NULL THEN
                JSONB_BUILD_OBJECT('name', person, 'id', person_id)
            WHEN person IS NOT NULL THEN
                JSONB_BUILD_OBJECT('name', person)
            ELSE
                JSONB_BUILD_OBJECT('name', 'N/A')
            END ORDER BY person) FILTER (WHERE person IS NOT NULL
            OR person_id IS NOT NULL) AS nominees
    FROM
        award a
    WHERE
        a.year = $1
        AND a.type = $2
    GROUP BY
        a.imdb_id,
        a.name,
        a.detail,
        a.winner
),
movie_awards AS (
    SELECT
        m.id AS movie_id,
        m.title,
        JSONB_AGG(JSONB_BUILD_OBJECT('winner', n.winner, 'category', n.category, 'detail', n.detail, 'nominees', COALESCE(n.nominees, '[]'::jsonb))
        ORDER BY n.winner DESC, n.category ASC) AS awards
    FROM
        movie m
        JOIN nominees n ON m.imdb_id = n.imdb_id
    GROUP BY
        m.id,
        m.title
)
SELECT
    *
FROM
    movie_awards
ORDER BY
    title ASC
	`

const awardsGroupedByCategoryQuery = `
SELECT
    a.name AS category,
    jsonb_agg(
        CASE WHEN person IS NOT NULL
            AND person_id IS NOT NULL THEN
            JSONB_BUILD_OBJECT('person', person, 'person_id', person_id, 'winner', winner, 'detail', detail, 'movie_id', m.id, 'title', m.title)
        WHEN person IS NOT NULL THEN
            JSONB_BUILD_OBJECT('person', person, 'winner', winner, 'detail', detail, 'movie_id', m.id, 'title', m.title)
        WHEN person IS NULL THEN
            JSONB_BUILD_OBJECT('movie_id', m.id, 'title', m.title, 'detail', detail)
        ELSE
            JSONB_BUILD_OBJECT('person', 'N/A')
        END ORDER BY person, title ASC) AS nominees
FROM
    award a
    LEFT JOIN movie m ON m.imdb_id = a.imdb_id
WHERE
    YEAR = $1
    AND type = $2
GROUP BY
    1
ORDER BY
    1
	`
