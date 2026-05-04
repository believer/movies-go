package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type RatingsHandler struct {
	repo db.RatingsQuerier
}

func NewRatingsHandler(repo db.RatingsQuerier) *RatingsHandler {
	return &RatingsHandler{repo}
}

func (h *RatingsHandler) GetMoviesByRating(c *fiber.Ctx) error {
	q := db.MakeQueries(c)
	rating, err := c.ParamsInt("rating")

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid rating")
	}

	movies, err := h.repo.GetMoviesByRating(q.UserID, rating, q.Offset)

	if err != nil {
		return fiber.ErrInternalServerError
	}

	// When there are no more movies to show, just return 200. Otherwise we
	// would display the "No movies seen" empty state which should only be
	// shown at the start.
	if len(movies) == 0 && q.Page > 1 {
		return c.SendStatus(fiber.StatusOK)
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this rating",
		Name:       fmt.Sprintf("Movies rated %d", rating),
		NextPage:   fmt.Sprintf("/rating/%d?page=%d", rating, q.Page+1),
		Movies:     movies,
	}))
}
