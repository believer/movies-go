package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func HandleGetSeries(c *fiber.Ctx) error {
	var series types.Series
	var movies []types.SeriesMovies

	id := utils.SelfHealingUrl(c.Params("id"))
	userId := c.Locals("UserId")

	// Get series information
	err := db.Dot.Get(db.Client, &series, "series-by-id", id)

	if err != nil {
		// TODO: Handle 404
		if err != sql.ErrNoRows {
			return err
		}
	}

	// Get series movies
	err = db.Dot.Select(db.Client, &movies, "series-movies-by-id", id, userId)

	if err != nil {
		// TODO: Handle 404
		if err != sql.ErrNoRows {
			return err
		}
	}

	totalMovies := 0

	for _, s := range movies {
		totalMovies += len(s.Movies)
	}

	props := views.SeriesProps{
		TotalMovies: totalMovies,
		Movies:      movies,
		Series:      series,
	}

	return utils.TemplRender(c, views.Series(props))
}
