package handlers

import (
	"believer/movies/components/list"
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"

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

	return utils.Render(c, views.ListView(views.ListViewProps{
		ListStyle:     list.Numbered,
		Name:          listData.Name,
		NumberColumns: 3,
		Movies:        movies,
	}))
}
