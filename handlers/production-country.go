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

type ProductionCountryHandler struct {
	repo db.ProductionCountryQuerier
}

func NewProductionCountryHandler(repo db.ProductionCountryQuerier) *ProductionCountryHandler {
	return &ProductionCountryHandler{repo}
}

func (h *ProductionCountryHandler) ListProductionCountries(c *fiber.Ctx) error {
	countries, err := h.repo.ListProductionCountries()

	if err != nil {
		return err
	}

	return utils.Render(c, views.RootView(views.RootViewProps{
		EmptyState: "No production countries",
		Title:      "Production countries",
		Items:      views.ToViewItems(countries),
	}))
}

func (h *ProductionCountryHandler) GetProductionCountry(c *fiber.Ctx) error {
	var movies types.Movies

	q := db.MakeQueries(c)
	country, err := h.repo.GetProductionCountryName(q.Id)

	if err != nil {
		return err
	}

	movies, err = h.repo.GetProductionCountryMovies(q.Id, q.UserID, q.Offset)

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

func (h *ProductionCountryHandler) GetProductionCountryStats(c *fiber.Ctx) error {
	q := db.MakeQueries(c)
	productionCountries, err := h.repo.GetProductionCountryStats(q.UserID, q.Year)

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
