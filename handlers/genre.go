package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type GenreHandler struct {
	repo db.GenreQuerier
}

func NewGenreHandler(repo db.GenreQuerier) *GenreHandler {
	return &GenreHandler{repo}
}

func (h *GenreHandler) ListGenres(c *fiber.Ctx) error {
	genres, err := h.repo.ListGenres()

	if err != nil {
		return err
	}

	return utils.Render(c, views.RootView(views.RootViewProps{
		EmptyState: "No genres",
		Title:      "Genres",
		Items:      views.ToViewItems(genres),
	}))
}

func (h *GenreHandler) GetGenre(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	id := req.IDString()
	genre, err := h.repo.GetGenreName(id)

	if err != nil {
		return err
	}

	movies, err := h.repo.GetGenreMovies(id, req.UserID(), req.Offset())

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
		EmptyState: "No movies for this genre",
		Name:       genre.Name,
		NextPage:   fmt.Sprintf("/genre/%s?page=%d", id, req.Page()+1),
		Movies:     movies,
	}))
}

func (h *GenreHandler) GetGenreStats(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	genres, err := h.repo.GetGenreStats(req.UserID(), req.Year())

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  genres,
		Title: "Genre",
		Href:  "/genre",
		Route: "/genre/stats",
		Root:  "genre",
		Year:  req.Year(),
		Years: req.AvailableYears(),
	}))
}
