package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func HandleFeed(c *fiber.Ctx) error {
	var movies types.Movies

	pageQuery := c.Query("page", "1")
	page, err := strconv.Atoi(pageQuery)

	if err != nil {
		page = 1
	}

	err = db.Client.Select(&movies, `
SELECT m.id, m.title, m.overview, m.release_date, s.date AS watched_at
FROM seen AS s
	INNER JOIN movie AS m ON m.id = s.movie_id
WHERE
	user_id = 1
ORDER BY s.date DESC
OFFSET $1
LIMIT 20
`, (page-1)*20)

	if err != nil {
		panic(err)
	}

	return c.Render("index", fiber.Map{
		"Movies":   movies,
		"NextPage": page + 1,
	})
}
