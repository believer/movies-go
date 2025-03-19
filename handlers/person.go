package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

func GetPersonByID(c *fiber.Ctx) error {
	var person types.Person
	var awards []types.Award

	id := utils.SelfHealingUrl(c.Params("id"))

	err := db.Dot.Get(db.Client, &person, "person-by-id", id, c.Locals("UserId"))

	if err != nil {
		// TODO: Display 404 page
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Person not found")
		}

		return err
	}

	err = db.Dot.Select(db.Client, &awards, "awards-by-person-id", id)

	if err != nil {
		return err
	}

	fields := []int{
		len(person.Cast),
		len(person.Director),
		len(person.Writer),
		len(person.Producer),
		len(person.Composer),
		len(person.Cinematographer),
		len(person.Editor),
	}

	totalCredits := 0
	for _, field := range fields {
		totalCredits += field
	}

	groupedAwards := make(map[string][]types.Award)

	won := 0
	for _, award := range awards {
		if award.Winner {
			won++
		}

		groupedAwards[award.Category] = append(groupedAwards[award.Category], award)
	}

	return utils.TemplRender(c, views.Person(views.PersonProps{
		Awards:       groupedAwards,
		Person:       person,
		TotalCredits: totalCredits,
		Won:          won,
	}))
}
