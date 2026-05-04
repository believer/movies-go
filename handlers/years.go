package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type YearsHandler struct {
	repo db.YearsQuerier
}

func NewYearsHandler(repo db.YearsQuerier) *YearsHandler {
	return &YearsHandler{repo}
}

func (h *YearsHandler) GetMoviesByYear(c *fiber.Ctx) error {
	q := db.MakeQueries(c)
	movies, err := h.repo.GetMoviesByYear(q.UserID, q.Year, q.Offset)

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies this year",
		NextPage:   fmt.Sprintf("/year/%s?page=%d", q.Year, q.Page+1),
		Movies:     movies,
		Name:       q.Year,
	}))
}
