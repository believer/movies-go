package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
)

type NowPlayingHandler struct {
	repo db.NowPlayingQuerier
}

func NewNowPlayingHandler(repo db.NowPlayingQuerier) *NowPlayingHandler {
	return &NowPlayingHandler{repo}
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
