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
	Id     string
	Page   int
	Offset int
	UserID string
	Year   string
	Years  []string
}

func MakeQueries(c *fiber.Ctx) *Queries {
	id := utils.SelfHealingUrlString(c.Params("id"))
	page := c.QueryInt("page", 1)
	userId := c.Locals("UserId").(string)
	year := c.Query("year", "All")
	years := append([]string{"All"}, utils.AvailableYears(time.Now())...)

	return &Queries{
		Id:     id,
		Page:   page,
		Offset: (page - 1) * 50,
		UserID: userId,
		Year:   year,
		Years:  years,
	}
}
