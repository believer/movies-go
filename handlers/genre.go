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

	page := c.QueryInt("page", 1)
	userId := c.Locals("UserId").(string)
	genreId := c.Params("id")

	err := db.Dot.Get(db.Client, &genre, "genre-by-id", genreId)

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &movies, "genres-by-id", genreId, userId, (page-1)*50)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, views.Genre(views.GenreProps{
		ID:       genreId,
		Name:     genre.Name,
		NextPage: page + 1,
		Movies:   movies,
	}))
}
