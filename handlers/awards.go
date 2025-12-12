package handlers

import (
	"believer/movies/db"
	"believer/movies/services/api"
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
	year := c.Params("year")
	sort := c.Query("sort", "Movie")

	a := api.New(c)

	switch sort {
	case "Movie":
		awards, err := a.AwardsByMovie(year)

		if err != nil {
			return err
		}

		return utils.Render(c, views.AwardsPage(views.AwardsPageProps{
			GroupedAwards: awards,
			Sort:          sort,
			Year:          year,
		}))
	case "Category":
		awards, err := a.AwardsByCategory(year)

		if err != nil {
			return err
		}

		return utils.Render(c, views.AwardsCategory(views.AwardsCategoryProps{
			Awards: awards,
			Sort:   sort,
			Year:   year,
		}))
	}

	return utils.Render(c, views.NotFound())
}
