package handlers

import (
	"believer/movies/components"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetGenre(c *fiber.Ctx) error {
	var movies types.Movies
	var genre types.MovieGenre

	page := c.QueryInt("page", 1)
	userId := c.Locals("UserId").(string)
	id, _ := utils.SelfHealingUrl(c.Params("id"))

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

	return utils.Render(c, components.ListView(components.ListViewProps{
		EmptyState: "No movies for this genre",
		Name:       genre.Name,
		NextPage:   fmt.Sprintf("/genre/%s?page=%d", id, page+1),
		Movies:     movies,
	}))
}
