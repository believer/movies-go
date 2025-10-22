package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetProductionCountry(c *fiber.Ctx) error {
	var country db.TableName
	var movies types.Movies

	q := db.MakeQueries(c)
	err := q.GetNameByID(&country, db.ProductionCountryTable)

	if err != nil {
		return err
	}

	err = q.GetMovies(&movies, db.ProductionCountryTable)

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this production country",
		Name:       country.Name,
		NextPage:   fmt.Sprintf("/production-country/%s?page=%d", q.Id, q.Page+1),
		Movies:     movies,
	}))
}

func GetProductionCountryStats(c *fiber.Ctx) error {
	var productionCountries []types.ListItem

	q := db.MakeQueries(c)
	err := q.GetStats(&productionCountries, db.ProductionCountryTable)

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
