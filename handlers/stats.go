package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"

	"github.com/gofiber/fiber/v2"
)

func HandleGetStats(c *fiber.Ctx) error {
	var stats struct {
		UniqueMovies      int `db:"unique_movies"`
		SeenWithRewatches int `db:"seen_with_rewatches"`
		TotalRuntime      int `db:"total_runtime"`
	}

	err := db.Client.Get(&stats, `
SELECT
	COUNT(DISTINCT movie_id) AS unique_movies,
	COUNT(movie_id) seen_with_rewatches,
  SUM(m.runtime) AS total_runtime
FROM
	seen AS s
INNER JOIN movie as m ON m.id = s.movie_id 
WHERE
	user_id = 1;
    `)

	if err != nil {
		return err
	}

	return c.Render("stats", fiber.Map{
		"Stats":                 stats,
		"FormattedTotalRuntime": utils.FormatRuntime(stats.TotalRuntime),
	})
}
