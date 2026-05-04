package db

import (
	"believer/movies/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

type TableName struct {
	Name string `db:"name"`
}

type Queries struct {
	Id              string
	IsAuthenticated bool
	Page            int
	Offset          int
	UserID          string
	Year            string
	Years           []string
}

func MakeQueries(c *fiber.Ctx) *Queries {
	id := utils.SelfHealingUrlString(c.Params("id"))
	year := c.Query("year", "All")
	page := c.QueryInt("page", 1)
	userId := c.Locals("UserId").(string)
	isAutheticated := c.Locals("IsAuthenticated").(bool)
	years := append([]string{"All"}, utils.AvailableYears(time.Now())...)

	return &Queries{
		Id:              id,
		IsAuthenticated: isAutheticated,
		Page:            page,
		Offset:          (page - 1) * 50,
		UserID:          userId,
		Year:            year,
		Years:           years,
	}
}
