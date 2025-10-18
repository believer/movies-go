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
	q := db.MakeProductionCountryQueries(c)

	country, err := q.ByID()

	if err != nil {
		return err
	}

	movies, err := q.Movies((page - 1) * 50)

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this production country",
		Name:       country.Name,
		NextPage:   fmt.Sprintf("/production-country/%s?page=%d", q.Id, page+1),
		Movies:     movies,
	}))
}

func GetProductionCountryStats(c *fiber.Ctx) error {
	q := db.MakeProductionCountryQueries(c)
	productionCountries, err := q.Stats()

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  productionCountries,
		Route: "/production-country/stats",
		Root:  "production-country",
		Title: "Production countries",
		Year:  q.Year,
		Years: q.Years,
	}))
}
