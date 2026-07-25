package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type SeriesHandler struct {
	repo db.SeriesQuerier
}

func NewSeriesHandler(repo db.SeriesQuerier) *SeriesHandler {
	return &SeriesHandler{repo}
}

func (h *SeriesHandler) GetSeries(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	id := req.IDString()

	// Get series information
	series, err := h.repo.GetSeriesByID(id)

	if err != nil {
		slog.Error("failed to get series", "error", err)
		// TODO: Handle 404
		if err != sql.ErrNoRows {
			return fiber.ErrInternalServerError
		}
	}

	// Get series movies
	movies, err := h.repo.GetSeriesMovies(id, req.UserID())

	if err != nil {
		slog.Error("failed to get series movies", "error", err)
		// TODO: Handle 404
		if err != sql.ErrNoRows {
			return err
		}
	}

	totalMovies, movies := calculateSeriesStats(movies)

	// Find the longest number in the series to set the correct
	// number of columns for the digits. Most times, a series doesn't
	// have more than one digit in the series, i.e., more than 9 movies.
	// But, there are cases, like MCU with more than 10 or Jackass with 2.5.
	longestNumber := 1

	for _, sm := range movies {
		for _, m := range sm.Movies {
			// -1 only includes the decimal if it's larger than zero
			seriesNumber := len(strconv.FormatFloat(m.NumberInSeries, 'f', -1, 64))
			longestNumber = max(seriesNumber, longestNumber)
		}
	}

	return utils.Render(c, views.Series(views.SeriesProps{
		LongestNumber: int(longestNumber),
		TotalMovies:   totalMovies,
		Movies:        movies,
		Series:        series,
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
