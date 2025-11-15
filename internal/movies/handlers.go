package movies

import (
	"believer/movies/components/movie"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Movie(c *fiber.Ctx) error {
	m, err := h.service.Movie(c)

	if err != nil {
		log.Error("Movie service error", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return utils.Render(c, views.Movie(
		views.MovieProps{
			Cast:          []views.CastDTO{},
			HasCharacters: true,
			WatchedAt:     []movie.WatchedAt{},
			IsInWatchlist: false,
			Movie:         m,
			Others:        types.OthersStats{},
			Review:        types.Review{},
			Back:          false,
		}))
}
