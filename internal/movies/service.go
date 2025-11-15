package movies

import (
	repo "believer/movies/internal/adapters/sqlc"

	"github.com/gofiber/fiber/v2"
)

type Service interface {
	Movie(ctx *fiber.Ctx) (repo.GetMovieRow, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{
		repo: repo,
	}
}

func (s *svc) Movie(c *fiber.Ctx) (repo.GetMovieRow, error) {
	id, err := c.ParamsInt("id")

	if err != nil {
		return repo.GetMovieRow{}, err
	}

	if id == 0 {
		id = c.QueryInt("id")
	}

	return s.repo.GetMovie(c.Context(), repo.GetMovieParams{
		ID:     int32(id),
		UserID: c.Locals("UserId").(int32),
	})
}
