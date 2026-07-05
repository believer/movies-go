package handlers

import (
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
	sort := req.QueryDefault("sort", "seen")
	l, err := h.repo.GetList(id)

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

	return utils.Render(c, views.ListPage(views.ListPageProps{
		Description: l.Description,
		Name:        fmt.Sprintf("%s - %s", l.Source, l.Name),
		Movies:      movies,
		Slug:        utils.CreateSelfHealingUrl("list", l.Slug, l.ID),
		Sort:        views.ToListSort(sort),
		Seen:        seen,
		Unseen:      len(movies) - seen,
	}))
}
