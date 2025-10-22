package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetMoviesByNumberOfAwards(c *fiber.Ctx) error {
	var movies types.Movies

	userId := c.Locals("UserId")
	numberOfAwards, err := c.ParamsInt("awards")
	includeNominations := c.QueryBool("nominations")

	if err != nil {
		return err
	}

	name := fmt.Sprintf("%d Academy Award wins", numberOfAwards)

	if includeNominations {
		name = fmt.Sprintf("%d Academy Award nominations", numberOfAwards)

		err = db.Client.Select(&movies, `
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
GROUP BY
    a.imdb_id,
    m.id,
    s.id
HAVING
    count(DISTINCT a.name) = $2
ORDER BY
    m.release_date DESC
			`, userId, numberOfAwards)

		if err != nil {
			return err
		}
	} else {

		err = db.Client.Select(&movies, `
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
GROUP BY
    a.imdb_id,
    m.id,
    s.id
HAVING
    count(DISTINCT a.name) = $2
ORDER BY
    m.release_date DESC
`, userId, numberOfAwards)

		if err != nil {
			return err
		}
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies with this amount of Academy Awards",
		Name:       name,
		Movies:     movies,
	}))
}

func GetAwardsByYear(c *fiber.Ctx) error {
	var awards []types.AwardsByYear

	year := c.Params("year")

	err := db.Client.Select(&awards, `WITH nominees AS (
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
		`, year)

	if err != nil {
		return err
	}

	return utils.Render(c, views.AwardsPage(views.AwardsPageProps{
		GroupedAwards: awards,
		Name:          year,
	}))
}
