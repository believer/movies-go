package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"database/sql"
	"time"

	"github.com/gofiber/fiber/v2"
)

func HandleGetMovieByID(c *fiber.Ctx) error {
	var movie types.Movie

	err := db.Client.Get(&movie, `
SELECT
	m.*,
  ARRAY_AGG(g.name) AS genres
FROM
	movie AS m
	INNER JOIN movie_genre AS mg ON mg.movie_id = m.id
	INNER JOIN genre AS g ON g.id = mg.genre_id
WHERE m.id = $1
GROUP BY 1
`, c.Params("id"))

	if err != nil {
		err := db.Client.Get(&movie, `
    SELECT
	m.*,
  ARRAY_AGG(g.name) AS genres
FROM
	movie AS m
	INNER JOIN movie_genre AS mg ON mg.movie_id = m.id
	INNER JOIN genre AS g ON g.id = mg.genre_id
-- Slugify function is defined in the database
WHERE slugify(m.title) ILIKE '%' || slugify($1) || '%'
GROUP BY 1
  `, c.Params("id"))

		if err != nil {
			// TODO: Handle this better
			if err == sql.ErrNoRows {
				return c.Status(404).SendString("Movie not found")
			}

			return err
		}
	}

	return c.Render("movie", fiber.Map{
		"Movie": movie,
	})
}

func HandleGetMovieCastByID(c *fiber.Ctx) error {
	var cast []types.Cast

	err := db.Client.Select(&cast, `
SELECT 
    INITCAP(mp.job::text) as job,
    JSONB_AGG(JSON_BUILD_OBJECT('name',p.name, 'id', p.id)) AS person
FROM 
    movie_person AS mp
    INNER JOIN person AS p ON p.id = mp.person_id
WHERE movie_id = $1
GROUP BY mp.job
ORDER BY
	CASE mp.job
		WHEN 'director' THEN 1
		WHEN 'writer' THEN 2
		WHEN 'cast' THEN 3
    WHEN 'composer' THEN 4
		WHEN 'producer' THEN 5
	END
`, c.Params("id"))

	if err != nil {
		panic(err)
	}

	return c.Render("partials/castList", fiber.Map{
		"Cast": cast,
	}, "")
}

func HandleGetMovieSeenByID(c *fiber.Ctx) error {
	var watchedAt []time.Time

	err := db.Client.Select(&watchedAt, `
SELECT date
FROM seen
WHERE movie_id = $1 AND user_id = 1
ORDER BY date DESC
`, c.Params("id"))

	if err != nil {
		panic(err)
	}

	return c.Render("partials/watched", fiber.Map{
		"WatchedAt": watchedAt,
	}, "")
}

// Render the add movie page
func HandleGetMovieNew(c *fiber.Ctx) error {
	return c.Render("add", nil)
}

// Handle adding a movies
func HandlePostMovieNew(c *fiber.Ctx) error {
	return c.SendString("TODO: Handle adding a movie")
}
