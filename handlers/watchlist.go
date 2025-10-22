package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
)

func released(userId, sortOrder string) (types.Movies, error) {
	var movies types.Movies

	err := db.Client.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.imdb_id,
    m.release_date,
    w.created_at
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date <= CURRENT_DATE
ORDER BY
    CASE WHEN $2 = 'Release date' THEN
        m.release_date
    ELSE
        w.created_at
    END ASC
		`, userId, sortOrder)

	return movies, err
}

func unreleased(userId, sortOrder string) (types.Movies, error) {
	var movies types.Movies

	err := db.Client.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.imdb_id,
    m.release_date,
    w.created_at
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date > CURRENT_DATE
ORDER BY
    CASE WHEN $2 = 'Date added' THEN
        w.created_at
    ELSE
        m.release_date
    END ASC
		`, userId, sortOrder)

	return movies, err
}

func tbd(userId string) (types.Movies, error) {
	var movies types.Movies

	err := db.Client.Select(&movies, `
SELECT
    m.id,
    m.title,
    m.imdb_id,
    m.release_date,
    w.created_at
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date IS NULL
		`, userId)

	return movies, err
}

func GetWatchlist(c *fiber.Ctx) error {
	userId := c.Locals("UserId").(string)
	movies, err := released(userId, "Date added")

	if err != nil {
		return err
	}

	unreleasedMovies, err := unreleased(userId, "Release date")

	if err != nil {
		return err
	}

	moviesWithoutReleaseDate, err := tbd(userId)

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
	sortOrder := c.Query("sortOrder", "Date added")
	userId := c.Locals("UserId").(string)

	var movies types.Movies

	movies, err := released(userId, sortOrder)

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
	sortOrder := c.Query("sortOrder", "Release date")
	userId := c.Locals("UserId").(string)

	movies, err := unreleased(userId, sortOrder)

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
