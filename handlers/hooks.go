package handlers

import (
	"believer/movies/db"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Progress struct {
	Completed bool   `form:"completed"`
	ImdbID    string `form:"imdb_id"`
	Name      string `json:"name"`
	Position  string `json:"position"`
}

func PlaybackProgress(c *fiber.Ctx) error {
	var data Progress

	if c.Locals("IsAuthenticated") == false {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	if err := c.BodyParser(&data); err != nil {
		return err
	}

	if data.ImdbID == "" {
		return c.SendStatus(fiber.StatusUnprocessableEntity)
	}

	if data.Completed {
		PostMovieNew(c)
	} else {
		// Convert string position to float
		positionParts := strings.Split(data.Position, ":")
		positionAsNumber := 0.0

		for i, p := range positionParts {
			n, err := strconv.Atoi(p)

			if err != nil {
				continue
			}

			switch i {
			case 0:
				positionAsNumber += 60 * float64(n)
			case 1:
				positionAsNumber += float64(n)
			case 2:
				positionAsNumber += float64(n) / 60
			}
		}

		_, err := db.Client.Exec(`
		INSERT INTO now_playing (imdb_id, position, user_id)
		    VALUES ($1, $2, $3)
		ON CONFLICT (imdb_id, user_id)
		    DO UPDATE SET
		        position = excluded.position
		`, data.ImdbID, positionAsNumber, c.Locals("UserId"))

		if err != nil {
			return err
		}
	}

	return c.SendStatus(200)
}
