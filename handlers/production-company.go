package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ProductionCompanyHandler struct {
	repo db.ProductionCompanyQuerier
}

func NewProductionCompanyHandler(repo db.ProductionCompanyQuerier) *ProductionCompanyHandler {
	return &ProductionCompanyHandler{repo}
}

func (h *ProductionCompanyHandler) ListProductionCompanies(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	page := req.Page()
	companies, err := h.repo.ListProductionCompanies(page)

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

func (h *ProductionCompanyHandler) GetProductionCompany(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	id := req.IDString()
	company, err := h.repo.GetProductionCompanyName(id)

	if err != nil {
		return err
	}

	movies, err := h.repo.GetProductionCompanyMovies(id, req.UserID(), req.Offset())

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this production company",
		Name:       company.Name,
		NextPage:   fmt.Sprintf("/production-company/%s?page=%d", id, req.Page()+1),
		Movies:     movies,
	}))
}

func (h *ProductionCompanyHandler) GetProductionCompanyStats(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	productionCompanies, err := h.repo.GetProductionCompanyStats(req.UserID(), req.Year())

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  productionCompanies,
		Href:  "/production-company",
		Route: "/production-company/stats",
		Root:  "production-company",
		Title: "Production companies",
		Year:  req.Year(),
		Years: req.AvailableYears(),
	}))
}
