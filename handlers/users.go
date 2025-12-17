package handlers

import (
	"believer/movies/components/list"
	"believer/movies/db"
	"believer/movies/utils"

	"github.com/gofiber/fiber/v2"
)

type User struct {
}

func GetUsers(c *fiber.Ctx) error {
	var options []list.DataListItem

	isAuth := utils.IsAuthenticated(c)
	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	err := db.Client.Select(&options, `
		SELECT
		    id AS "value",
		    name
		FROM
		    "user"
		`)

	if err != nil {
		return err
	}

	return utils.Render(c, list.DataList(options, "friend_list"))
}
