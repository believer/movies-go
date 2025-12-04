package handlers

import (
	"believer/movies/services/api"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
)

func NowPlaying(c *fiber.Ctx) error {
	api := api.New(c)
	nowPlaying, err := api.NowPlaying()

	if err != nil {
		return err
	}

	return utils.Render(c, views.NowPlaying(views.NowPlayingProps{
		Movies: nowPlaying,
	}))
}
