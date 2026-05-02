package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"database/sql"

	"github.com/gofiber/fiber/v2"
)

type PersonHandler struct {
	repo db.PersonQuerier
}

func NewPersonHandler(repo db.PersonQuerier) *PersonHandler {
	return &PersonHandler{repo}
}

func (h *PersonHandler) GetPersonByID(c *fiber.Ctx) error {
	q := db.MakeQueries(c)
	person, err := h.repo.GetPersonByID(q.Id, q.UserID)

	if err != nil {
		// TODO: Display 404 page
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString("Person not found")
		}

		return err
	}

	academyAwards, academyWins, academyOrder, err := h.repo.GetGroupedAwards(q.Id, db.AcademyAward)

	if err != nil {
		return err
	}

	baftas, baftaWins, baftaOrder, err := h.repo.GetGroupedAwards(q.Id, db.Bafta)

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

	return utils.Render(c, views.Person(views.PersonProps{
		AcademyAwards:      academyAwards,
		AcademyAwardsOrder: academyOrder,
		AcademyAwardsWon:   academyWins,
		Baftas:             baftas,
		BaftaOrder:         baftaOrder,
		BaftasWon:          baftaWins,
		Person:             person,
		TotalCredits:       totalCredits,
	}))
}
