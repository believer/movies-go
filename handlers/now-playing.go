package handlers

import (
	"believer/movies/db"
	"believer/movies/services/api"
	"believer/movies/utils"
	"believer/movies/views"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type NowPlayingHandler struct {
	repo      db.NowPlayingQuerier
	movieRepo db.MovieQuerier
}

func NewNowPlayingHandler(repo db.NowPlayingQuerier, movieRepo db.MovieQuerier) *NowPlayingHandler {
	return &NowPlayingHandler{repo, movieRepo}
}

func (h *NowPlayingHandler) GetNowPlaying(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	nowPlaying, err := h.repo.GetNowPlaying(req.UserID())

	if err != nil {
		return err
	}

	return utils.Render(c, views.NowPlaying(views.NowPlayingProps{
		Movies: nowPlaying,
	}))
}

func (h *NowPlayingHandler) GetManage(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()

	if !isAuth {
		return c.Redirect("/login")
	}

	nowPlaying, err := h.repo.GetNowPlaying(req.UserID())
	if err != nil {
		return err
	}

	return utils.Render(c, views.NowPlayingManage(views.NowPlayingManageProps{
		Movies: nowPlaying,
	}))
}

func (h *NowPlayingHandler) PostManage(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()
	userID := req.UserID()

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	imdbID := req.FormValue("imdb_id")
	positionStr := req.FormValue("position")
	position, err := strconv.ParseFloat(positionStr, 64)
	if err != nil {
		position = 0
	}

	imdbIDClean, err := utils.ParseId(imdbID)
	if err != nil {
		c.Set("HX-Retarget", "#error")
		return c.SendString(err.Error())
	}

	movieExists, err := h.movieRepo.MovieExists(imdbIDClean)
	if err != nil {
		return err
	}

	var movieID int
	if movieExists {
		movieObj, err := h.movieRepo.GetMovieByImdbID(imdbIDClean)
		if err != nil {
			return err
		}
		movieID = movieObj.ID
	} else {
		apiInstance := api.New(c)
		_, insertedMovieID, err := apiInstance.AddMovie(imdbIDClean, false)
		if err != nil {
			return err
		}
		movieID = insertedMovieID
	}

	err = h.movieRepo.UpdateNowPlaying(movieID, position, userID)
	if err != nil {
		return err
	}

	InvalidateStatsCache(userID)

	if req.Get("HX-Request") == "true" {
		c.Set("HX-Redirect", "/now-playing/manage")
		return c.SendStatus(fiber.StatusOK)
	}
	return c.Redirect("/now-playing/manage")
}

func (h *NowPlayingHandler) PutNowPlaying(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()
	userID := req.UserID()

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	movieIDStr := req.Params("id")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		return err
	}

	positionStr := req.FormValue("position")
	position, err := strconv.ParseFloat(positionStr, 64)
	if err != nil {
		return err
	}

	err = h.movieRepo.UpdateNowPlaying(movieID, position, userID)
	if err != nil {
		return err
	}

	InvalidateStatsCache(userID)

	movieObj, err := h.movieRepo.GetMovieByIDSimple(movieID)
	if err != nil {
		return err
	}
	movieObj.Position = position

	return utils.Render(c, views.NowPlayingManageItem(movieObj))
}

func (h *NowPlayingHandler) DeleteNowPlaying(c *fiber.Ctx) error {
	req := utils.NewRequest(c)
	isAuth := req.IsAuthenticated()
	userID := req.UserID()

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	movieIDStr := req.Params("id")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		return err
	}

	err = h.movieRepo.DeleteNowPlayingDirect(userID, movieID)
	if err != nil {
		return err
	}

	InvalidateStatsCache(userID)

	return c.SendString("")
}
