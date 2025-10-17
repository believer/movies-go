package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetProductionCountry(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	queries, _ := db.MakeProductionCountryQueries(c)

	country, err := queries.ByID()

	if err != nil {
		return err
	}

	movies, err := queries.Movies((page - 1) * 50)

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this production country",
		Name:       country.Name,
		NextPage:   fmt.Sprintf("/production-country/%s?page=%d", queries.Id, page+1),
		Movies:     movies,
	}))
}
