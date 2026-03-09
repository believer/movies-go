package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ProductionItem struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func (p ProductionItem) Title() string {
	return p.Name
}

func (p ProductionItem) Subtitle() string {
	return ""
}

func (p ProductionItem) Href() string {
	return utils.CreateSelfHealingUrl("production-country", p.Name, p.ID)
}

func ListProductionCountries(c *fiber.Ctx) error {
	var countries []ProductionItem

	err := db.Client.Select(&countries, `
		SELECT
		    id,
		    name
		FROM
		    production_country
		ORDER BY
		    name ASC
		`)

	if err != nil {
		return err
	}

	return utils.Render(c, views.RootView(views.RootViewProps{
		EmptyState: "No production countries",
		Title:      "Production countries",
		Items:      views.ToViewItems(countries),
	}))
}

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
		Href:  "/production-country",
		Route: "/production-country/stats",
		Root:  "production-country",
		Title: "Production countries",
		Year:  q.Year,
		Years: q.Years,
	}))
}
