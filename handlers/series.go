package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type SeriesHandler struct {
	repo db.SeriesQuerier
}

func NewSeriesHandler(repo db.SeriesQuerier) *SeriesHandler {
	return &SeriesHandler{repo}
}

func (h *SeriesHandler) GetSeries(c *fiber.Ctx) error {
	q := db.MakeQueries(c)

	// Get series information
	series, err := h.repo.GetSeriesByID(q.Id)

	if err != nil {
		slog.Error("failed to get series", "error", err)
		// TODO: Handle 404
		if err != sql.ErrNoRows {
			return fiber.ErrInternalServerError
		}
	}

	// Get series movies
	movies, err := h.repo.GetSeriesMovies(q.Id, q.UserID)

	if err != nil {
		slog.Error("failed to get series movies", "error", err)
		// TODO: Handle 404
		if err != sql.ErrNoRows {
			return err
		}
	}

	totalMovies, movies := calculateSeriesStats(movies)

	return utils.Render(c, views.Series(views.SeriesProps{
		TotalMovies: totalMovies,
		Movies:      movies,
		Series:      series,
	}))
}

func calculateSeriesStats(movies []types.SeriesMovies) (total int, withSeenCounts []types.SeriesMovies) {
	for i, s := range movies {
		seen := 0
		total += len(s.Movies)

		for _, m := range s.Movies {
			if m.Seen {
				seen++
			}
		}

		movies[i].Seen = seen
	}

	return total, movies
}
