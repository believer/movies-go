package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
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
	public.movie AS m
	INNER JOIN public.movie_genre AS mg ON mg.movie_id = m.id
	INNER JOIN public.genre AS g ON g.id = mg.genre_id
WHERE m.id = $1
GROUP BY 1
`, c.Params("id"))

	if err != nil {
		panic(err)
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
    public.movie_person AS mp
    INNER JOIN public.person AS p ON p.id = mp.person_id
WHERE movie_id = $1
GROUP BY mp.job
ORDER BY
	CASE mp.job
		WHEN 'cast' THEN 1
		WHEN 'director' THEN 2
    WHEN 'composer' THEN 3
		WHEN 'writer' THEN 4
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
FROM public.seen
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
