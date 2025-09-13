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
	includeNominations := c.QueryBool("nominations")

	if err != nil {
		return err
	}

	name := fmt.Sprintf("%d Academy Award wins", numberOfAwards)

	if includeNominations {
		name = fmt.Sprintf("%d Academy Award nominations", numberOfAwards)

		err = db.Dot.Select(db.Client, &movies, "movies-by-number-of-nominations", userId, numberOfAwards)

		if err != nil {
			return err
		}
	} else {

		err = db.Dot.Select(db.Client, &movies, "movies-by-number-of-wins", userId, numberOfAwards)

		if err != nil {
			return err
		}
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies with this amount of Academy Awards",
		Name:       name,
		Movies:     movies,
	}))
}

func GetAwardsByYear(c *fiber.Ctx) error {
	var awards []types.GlobalAward

	year := c.Params("year")

	err := db.Dot.Select(db.Client, &awards, "awards-by-year", year)

	if err != nil {
		return err
	}

	return utils.Render(c, views.AwardsPage(views.AwardsPageProps{
		GroupedAwards: awards,
		Name:          year,
	}))
}
