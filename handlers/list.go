package handlers

import (
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
	req := utils.NewRequest(c)
	listData, err := h.repo.GetLists(req.UserID())

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
		Title:       l.Name,
		Subtitle:    l.Source,
		Movies:      movies,
		Slug:        utils.CreateSelfHealingUrl("list", l.Slug, l.ID),
		Sort:        views.ToListSort(sort),
		Seen:        seen,
		Unseen:      len(movies) - seen,
	}))
}

func (h *ListHandler) GetListsByMovieId(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	movieID := req.IDString()

	lists, err := h.repo.GetListsByMovieID(movieID)

	if err != nil {
		return err
	}

	return utils.Render(c, views.MovieLists(views.MovieListsProps{
		Lists: lists,
	}))
}
