package handlers

import (
	"believer/movies/db"
	"believer/movies/types"

	"github.com/gofiber/fiber/v2"
)

func HandleMovieSearch(c *fiber.Ctx) error {
	var movies types.Movies

	search := c.FormValue("search")

	if search == "" {
		return c.Redirect("/")
	}

	err := db.Client.Select(&movies, `
SELECT m.id, m.title, m.overview, m.release_date, s.date AS watched_at
FROM public.seen AS s
	INNER JOIN public.movie AS m ON m.id = s.movie_id
WHERE
	user_id = 1
	AND m.title ILIKE '%' || $1 || '%'
ORDER BY s.date DESC
`, search)

	if err != nil {
		panic(err)
	}

	return c.Render("index", fiber.Map{
		"Movies": movies,
	})
}
