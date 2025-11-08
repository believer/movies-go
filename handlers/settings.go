package handlers

import (
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
)

func Settings(c *fiber.Ctx) error {
	return utils.Render(c, views.Settings())
}
