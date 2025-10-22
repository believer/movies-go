package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
)

func GetWatchlist(c *fiber.Ctx) error {
	var movies types.Movies
	var unreleasedMovies types.Movies
	var moviesWithoutReleaseDate types.Movies

	userId := c.Locals("UserId")

	err := db.Dot.Select(db.Client, &movies, "watchlist", userId, "Date added")

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &unreleasedMovies, "watchlist-unreleased", userId, "Release date")

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &moviesWithoutReleaseDate, "watchlist-no-date", userId)

	if err != nil {
		return err
	}

	return utils.Render(c, views.Watchlist(views.WatchlistProps{
		Movies:                   movies,
		UnreleasedMovies:         unreleasedMovies,
		MoviesWithoutReleaseDate: moviesWithoutReleaseDate,
	}))
}

func GetWatchlistMovies(c *fiber.Ctx) error {
	var movies types.Movies

	sortOrder := c.Query("sortOrder", "Date added")
	userId := c.Locals("UserId")

	err := db.Dot.Select(db.Client, &movies, "watchlist", userId, sortOrder)

	if err != nil {
		return err
	}

	return utils.Render(c, views.WatchlistList(
		views.WatchlistListProps{
			Movies: movies,
			Title:  "Movies",
			Action: views.SortWatchlist("/watchlist/movies", sortOrder),
			Order:  sortOrder,
		}))
}

func GetWatchlistUnreleasedMovies(c *fiber.Ctx) error {
	var movies types.Movies

	sortOrder := c.Query("sortOrder", "Release date")
	userId := c.Locals("UserId")

	err := db.Dot.Select(db.Client, &movies, "watchlist-unreleased", userId, sortOrder)

	if err != nil {
		return err
	}

	return utils.Render(c, views.WatchlistList(
		views.WatchlistListProps{
			Movies: movies,
			Title:  "Unreleased movies",
			Action: views.SortWatchlist("/watchlist/unreleased-movies", sortOrder),
			Order:  sortOrder,
		}))
}

func DeleteFromWatchlist(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieId := c.Params("id")
	userId := c.Locals("UserId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	_, err := db.Client.Exec(`
DELETE FROM watchlist
WHERE movie_id = $1
    AND user_id = $2
		`, movieId, userId)

	if err != nil {
		return err
	}

	return c.SendString("")
}
