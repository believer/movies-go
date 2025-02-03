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
	id := utils.SelfHealingUrl(c.Params("id"))

	err := db.Dot.Get(db.Client, &genre, "genre-by-id", id)

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &movies, "movies-by-genre-id", id, userId, (page-1)*50)

	if err != nil {
		return err
	}

	// When there are no more movies to show, just return 200. Otherwise we
	// would display the "No movies seen" empty state which should only be
	// shown at the start.
	if len(movies) == 0 && page > 1 {
		return c.SendStatus(fiber.StatusOK)
	}

	return utils.TemplRender(c, views.Genre(views.GenreProps{
		ID:       id,
		Name:     genre.Name,
		NextPage: page + 1,
		Movies:   movies,
	}))
}
