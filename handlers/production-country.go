package handlers

import (
	"believer/movies/db"
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
	req := utils.NewRequest(c)
	id := req.IDString()
	country, err := h.repo.GetProductionCountryName(id)

	if err != nil {
		return err
	}

	movies, err := h.repo.GetProductionCountryMovies(id, req.UserID(), req.Offset())

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this production country",
		Name:       country.Name,
		NextPage:   fmt.Sprintf("/production-country/%s?page=%d", id, req.Page()+1),
		Movies:     movies,
	}))
}

func (h *ProductionCountryHandler) GetProductionCountryStats(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	productionCountries, err := h.repo.GetProductionCountryStats(req.UserID(), req.Year())

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  productionCountries,
		Href:  "/production-country",
		Route: "/production-country/stats",
		Root:  "production-country",
		Title: "Production countries",
		Year:  req.Year(),
		Years: req.AvailableYears(),
	}))
}
