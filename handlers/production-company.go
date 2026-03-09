package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func ListProductionCompanies(c *fiber.Ctx) error {
	var companies []ProductionItem

	page := c.QueryInt("page", 1)

	err := db.Client.Select(&companies, `
		SELECT
		    id,
		    name
		FROM
		    production_company
		ORDER BY
		    name ASC OFFSET $1
		LIMIT 50
		`, (page-1)*50)

	if err != nil {
		return err
	}

	return utils.Render(c, views.RootView(views.RootViewProps{
		EmptyState: "No production companies",
		NextPage:   fmt.Sprintf("/production-company?page=%d", page+1),
		Title:      "Production companies",
		Items:      views.ToViewItems(companies),
	}))
}

func GetProductionCompany(c *fiber.Ctx) error {
	var company db.TableName
	var movies types.Movies

	q := db.MakeQueries(c)
	err := q.GetNameByID(&company, db.ProductionCompanyTable)

	if err != nil {
		return err
	}

	err = q.GetMovies(&movies, db.ProductionCompanyTable)

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this production company",
		Name:       company.Name,
		NextPage:   fmt.Sprintf("/production-company/%s?page=%d", q.Id, q.Page+1),
		Movies:     movies,
	}))
}

func GetProductionCompanyStats(c *fiber.Ctx) error {
	var productionCompanies []types.ListItem

	q := db.MakeQueries(c)
	err := q.GetStats(&productionCompanies, db.ProductionCompanyTable)

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  productionCompanies,
		Href:  "/production-company",
		Route: "/production-company/stats",
		Root:  "production-company",
		Title: "Production companies",
		Year:  q.Year,
		Years: q.Years,
	}))
}
