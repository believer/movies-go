package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func HandleGetPersonByID(c *fiber.Ctx) error {
	var person types.Person

	id := utils.SelfHealingUrl(c.Params("id"))

	err := db.Dot.Get(db.Client, &person, "person-by-id", id, c.Locals("UserId"))

	if err != nil {
		// TODO: Display 404 page
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Person not found")
		}

		return err
	}

	fields := []int{
		len(person.Cast),
		len(person.Director),
		len(person.Writer),
		len(person.Producer),
		len(person.Composer),
	}

	totalCredits := 0
	for _, field := range fields {
		totalCredits += field
	}

	return utils.TemplRender(c, views.Person(person, totalCredits, id))
}
