package db

import (
	"believer/movies/types"
	"believer/movies/utils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lib/pq"
)

type Relation struct {
	Table        string
	ToMovieTable string
	Column       string
}

type TableName struct {
	Name string `db:"name"`
}

var (
	GenreTable             = Relation{"genre", "movie_genre", "t.genre_id"}
	LanguageTable          = Relation{"language", "movie_language", "t.language_id"}
	ProductionCompanyTable = Relation{"production_company", "movie_company", "t.company_id"}
	ProductionCountryTable = Relation{"production_country", "movie_country", "t.country_id"}
)

type Queries struct {
	Id     string
	Page   int
	Offset int
	UserId string
	Year   string
	Years  []string
}

func MakeQueries(c *fiber.Ctx) *Queries {
	id := utils.SelfHealingUrlString(c.Params("id"))
	page := c.QueryInt("page", 1)
	userId := c.Locals("UserId").(string)
	year := c.Query("year", "All")
	years := append([]string{"All"}, utils.AvailableYears()...)

	return &Queries{
		Id:     id,
		Page:   page,
		Offset: (page - 1) * 50,
		UserId: userId,
		Year:   year,
		Years:  years,
	}
}

// Get name of an entity
func (q *Queries) GetNameByID(dest *TableName, relation Relation) error {
	query := strings.ReplaceAll(`SELECT
    name
FROM
    {{table}}
WHERE
    id = $1`, "{{table}}", pq.QuoteIdentifier(relation.Table))

	return Client.Get(dest, query, q.Id)
}

// Get all movies, paginated, for a specific entity
func (q *Queries) GetMovies(dest *types.Movies, relation Relation) error {
	query := `
		SELECT DISTINCT
		    (m.id),
		    m.title,
		    m.release_date,
		    (s.id IS NOT NULL) AS "seen"
		FROM
		    {{table}} AS t
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
		    {{column}} = $1
		ORDER BY
		    m.release_date DESC OFFSET $3
		LIMIT 50
	`
	query = strings.ReplaceAll(query, "{{table}}", pq.QuoteIdentifier(relation.ToMovieTable))
	query = strings.ReplaceAll(query, "{{column}}", relation.Column)

	return Client.Select(dest, query, q.Id, q.UserId, q.Offset)
}

// Get top 10 of a specific entity for the stats page
func (q *Queries) GetStats(dest *[]types.ListItem, relation Relation) error {
	query := `
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
    INNER JOIN {{relation_table}} t ON t.movie_id = s.movie_id
    INNER JOIN {{table}} pc ON pc.id = {{column}}
GROUP BY
    pc.id
ORDER BY
    count DESC
LIMIT 10
	`

	query = strings.ReplaceAll(query, "{{table}}", pq.QuoteIdentifier(relation.Table))
	query = strings.ReplaceAll(query, "{{relation_table}}", pq.QuoteIdentifier(relation.ToMovieTable))
	query = strings.ReplaceAll(query, "{{column}}", relation.Column)

	return Client.Select(dest, query, q.UserId, q.Year)
}
