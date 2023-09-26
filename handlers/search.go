package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func HandleMovieSearch(c *fiber.Ctx) error {
	var movies types.Movies

	search := c.FormValue("search")

	if search == "" {
		return c.Redirect("/")
	}

	err := db.Client.Select(&movies, `
SELECT m.id, m.title, m.overview, m.release_date AS watched_at
FROM movie AS m
WHERE m.title ILIKE '%' || $1 || '%'
ORDER BY m.release_date DESC
`, search)

	if err != nil {
		// TODO: Display 404 page
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Movie not found")
		}

		return err
	}

	return c.Render("index", fiber.Map{
		"Movies": movies,
	})
}
