package handlers

import (
	"believer/movies/components/list"
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ListHandler struct {
	repo db.ListQuerier
}

func NewListHandler(repo db.ListQuerier) *ListHandler {
	return &ListHandler{repo}
}

func (h *ListHandler) GetLists(c *fiber.Ctx) error {
	listData, err := h.repo.GetLists()

	if err != nil {
		return utils.Render(c, views.NotFound())
	}

	return utils.Render(c, views.RootView(views.RootViewProps{
		EmptyState: "No lists",
		Title:      "Official lists",
		Items:      views.ToViewItems(listData),
	}))
}

func (h *ListHandler) GetListById(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	id := req.IDString()
	listData, err := h.repo.GetList(id)

	if err != nil {
		return utils.Render(c, views.NotFound())
	}

	movies, err := h.repo.GetListMovies(id, req.UserID())

	if err != nil {
		return utils.Render(c, views.NotFound())
	}

	seen := 0
	for _, m := range movies {
		if m.Seen {
			seen += 1
		}
	}
	percentage := (float64(seen) / float64(len(movies))) * 100

	return utils.Render(c, views.ListView(views.ListViewProps{
		Completion:    fmt.Sprintf("Completed %.0f%% (%d / %d)", percentage, seen, len(movies)),
		Description:   listData.Description,
		EmptyState:    "No movies in list",
		ListStyle:     list.Numbered,
		Name:          fmt.Sprintf("%s - %s", listData.Source, listData.Name),
		NumberColumns: 3,
		Movies:        movies,
	}))
}
