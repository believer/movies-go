package handlers

import (
	"believer/movies/components"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetMoviesByNumberOfAwards(c *fiber.Ctx) error {
	var movies types.Movies

	userId := c.Locals("UserId")
	numberOfAwards, err := c.ParamsInt("awards")

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &movies, "movies-by-number-of-awards", userId, numberOfAwards)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.ListView(components.ListViewProps{
		EmptyState: "No movies with this amount of Academy Awards",
		Name:       fmt.Sprintf("Won %d Academy Awards", numberOfAwards),
		Movies:     movies,
	}))
}
