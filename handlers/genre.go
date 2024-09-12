package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
)

func HandleGetGenre(c *fiber.Ctx) error {
	var movies types.Movies
	var genre types.MovieGenre

	userId := c.Locals("UserId").(string)
	genreId := c.Params("id")

	err := db.Dot.Get(db.Client, &genre, "genre-by-id", genreId)

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &movies, "genres-by-id", genreId, userId)

	if err != nil {
		return err
	}

	seen := 0

	for _, movie := range movies {
		if movie.Seen {
			seen += 1
		}
	}

	return utils.TemplRender(c, views.Genre(views.GenreProps{
		Name:   genre.Name,
		Movies: movies,
		Seen:   seen,
	}))
}
