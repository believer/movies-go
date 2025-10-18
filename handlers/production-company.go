package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetProductionCompany(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	q := db.MakeProductionCompanyQueries(c)

	company, err := q.ByID()

	if err != nil {
		return err
	}

	movies, err := q.Movies((page - 1) * 50)

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this production company",
		Name:       company.Name,
		NextPage:   fmt.Sprintf("/production-company/%s?page=%d", q.Id, page+1),
		Movies:     movies,
	}))
}

func GetProductionCompanyStats(c *fiber.Ctx) error {
	q := db.MakeProductionCompanyQueries(c)
	productionCompanies, err := q.Stats()

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  productionCompanies,
		Route: "/production-company/stats",
		Root:  "production-company",
		Title: "Production companies",
		Year:  q.Year,
		Years: q.Years,
	}))
}
