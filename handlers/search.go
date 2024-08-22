package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func HandleMovieSearch(c *fiber.Ctx) error {
	var movies types.Movies

	pageQuery := c.Query("page", "1")
	page, err := strconv.Atoi(pageQuery)

	if err != nil {
		page = 1
	}
	search := c.FormValue("search")

	if search == "" {
		return c.Redirect("/")
	}

	err = db.Client.Select(&movies, `
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

	return utils.TemplRender(c, views.Feed(
		utils.IsAuthenticated(c),
		movies,
		page+1,
		search,
	))
}
