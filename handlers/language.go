package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type LanguageHandler struct {
	repo db.LanguageQuerier
}

func NewLanguageHandler(repo db.LanguageQuerier) *LanguageHandler {
	return &LanguageHandler{repo}
}

func (h *LanguageHandler) ListLanguages(c *fiber.Ctx) error {
	languages, err := h.repo.ListLanguages()

	if err != nil {
		return err
	}

	return utils.Render(c, views.RootView(views.RootViewProps{
		EmptyState: "No languages",
		Title:      "Languages",
		Items:      views.ToViewItems(languages),
	}))
}

func (h *LanguageHandler) GetLanguage(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	id := req.IDString()
	language, err := h.repo.GetLanguageName(id)

	if err != nil {
		return err
	}

	movies, err := h.repo.GetLanguageMovies(id, req.UserID(), req.Offset())

	if err != nil {
		return err
	}

	// When there are no more movies to show, just return 200. Otherwise we
	// would display the "No movies seen" empty state which should only be
	// shown at the start.
	if len(movies) == 0 && req.Page() > 1 {
		return c.SendStatus(fiber.StatusOK)
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this language",
		Name:       language.Name,
		NextPage:   fmt.Sprintf("/language/%s?page=%d", id, req.Page()+1),
		Movies:     movies,
	}))
}

func (h *LanguageHandler) GetLanguageStats(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	languages, err := h.repo.GetLanguageStats(req.UserID(), req.Year())

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  languages,
		Title: "Language",
		Root:  "language",
		Href:  "/language",
		Route: "/language/stats",
		Year:  req.Year(),
		Years: req.AvailableYears(),
	}))
}
