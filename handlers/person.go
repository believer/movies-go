package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func HandleGetPersonByID(c *fiber.Ctx) error {
	var person types.Person

	err := db.Dot.Get(db.Client, &person, "person-by-id", c.Params("id"))

	if err != nil {
		// TODO: Display 404 page
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Person not found")
		}

		return err
	}

	return c.Render("person", fiber.Map{
		"Person": person,
	})
}
