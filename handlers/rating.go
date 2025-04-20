package handlers

import (
	"believer/movies/components"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetMoviesByRating(c *fiber.Ctx) error {
	var movies types.Movies

	page := c.QueryInt("page", 1)
	userId := c.Locals("UserId")
	rating, err := c.ParamsInt("rating")

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &movies, "movies-by-rating", rating, userId, (page-1)*50)

	if err != nil {
		return err
	}

	// When there are no more movies to show, just return 200. Otherwise we
	// would display the "No movies seen" empty state which should only be
	// shown at the start.
	if len(movies) == 0 && page > 1 {
		return c.SendStatus(fiber.StatusOK)
	}

	return utils.TemplRender(c, components.ListView(components.ListViewProps{
		EmptyState: "No movies for this rating",
		Name:       fmt.Sprintf("Movies rated %d", rating),
		NextPage:   fmt.Sprintf("/rating/%d?page=%d", rating, page+1),
		Movies:     movies,
	}))
}
