package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
)

type WatchlistHandler struct {
	repo db.WatchlistQuerier
}

func NewWatchlistHandler(repo db.WatchlistQuerier) *WatchlistHandler {
	return &WatchlistHandler{repo}
}

func (h *WatchlistHandler) GetWatchlist(c *fiber.Ctx) error {
	q := db.MakeQueries(c)
	movies, err := h.repo.GetReleasedMovies(q.UserID, "Date added")

	if err != nil {
		return err
	}

	unreleasedMovies, err := h.repo.GetUnreleasedMovies(q.UserID, "Release date")

	if err != nil {
		return err
	}

	moviesWithoutReleaseDate, err := h.repo.GetTBDMovies(q.UserID)

	if err != nil {
		return err
	}

	return utils.Render(c, views.Watchlist(views.WatchlistProps{
		Movies:                   movies,
		UnreleasedMovies:         unreleasedMovies,
		MoviesWithoutReleaseDate: moviesWithoutReleaseDate,
	}))
}

func (h *WatchlistHandler) GetWatchlistMovies(c *fiber.Ctx) error {
	sortOrder := c.Query("sortOrder", "Date added")
	q := db.MakeQueries(c)

	movies, err := h.repo.GetReleasedMovies(q.UserID, sortOrder)

	if err != nil {
		return err
	}

	return utils.Render(c, views.WatchlistList(
		views.WatchlistListProps{
			Movies:      movies,
			Title:       "Movies",
			Action:      views.SortWatchlist("/watchlist/movies", sortOrder),
			Order:       sortOrder,
			ReleaseDate: views.Released,
		}))
}

func (h *WatchlistHandler) GetWatchlistUnreleasedMovies(c *fiber.Ctx) error {
	sortOrder := c.Query("sortOrder", "Release date")
	q := db.MakeQueries(c)

	movies, err := h.repo.GetUnreleasedMovies(q.UserID, sortOrder)

	if err != nil {
		return err
	}

	return utils.Render(c, views.WatchlistList(
		views.WatchlistListProps{
			Movies:      movies,
			Title:       "Unreleased movies",
			Action:      views.SortWatchlist("/watchlist/unreleased-movies", sortOrder),
			Order:       sortOrder,
			ReleaseDate: views.Unreleased,
		}))
}

func (h *WatchlistHandler) DeleteFromWatchlist(c *fiber.Ctx) error {
	q := db.MakeQueries(c)

	if !q.IsAuthenticated {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	err := h.repo.DeleteFromWatchlist(q.Id, q.UserID)

	if err != nil {
		return err
	}

	return c.SendString("")
}
