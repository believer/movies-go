package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
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

	return utils.TemplRender(c, views.Awards(views.AwardsProps{
		Name:   fmt.Sprintf("Won %d Academy Awards", numberOfAwards),
		Movies: movies,
	}))
}
