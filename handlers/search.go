package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"
	"strconv"
	"strings"
	"time"

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

	if strings.Contains(c.Get("Accept"), "hyperview") {
		lastMovieYear := c.Query("lastMovieYear", movies[len(movies)-1].WatchedAt.Format("2006"))

		return c.Render("feed_pages", fiber.Map{
			"Movies":        movies,
			"Page":          page + 1,
			"LastMovieYear": lastMovieYear,
			"CurrentYear":   time.Now().Year(),
			"SearchQuery":   search,
		})
	}

	return utils.TemplRender(c, views.Feed(
		utils.IsAuthenticated(c),
		movies,
		page+1,
	))
}
