package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
)

func HandleGetWatchlist(c *fiber.Ctx) error {
	var movies types.Movies
	var unreleasedMovies types.Movies

	userId := c.Locals("UserId")

	err := db.Dot.Select(db.Client, &movies, "watchlist", userId)

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &unreleasedMovies, "watchlist-unreleased", userId)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, views.Watchlist(views.WatchlistProps{
		Movies:           movies,
		UnreleasedMovies: unreleasedMovies,
	}))
}
