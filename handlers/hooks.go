package handlers

import (
	"believer/movies/db"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Progress struct {
	Position string `json:"position"`
	ImdbID   string `json:"imdb_id"`
	Name     string `json:"name"`
}

func ProgressHook(c *fiber.Ctx) error {
	if c.Locals("IsAuthenticated") == nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	var progress Progress
	if err := json.Unmarshal(c.Body(), &progress); err != nil {
		return err
	}

	// Convert string position to float
	positionParts := strings.Split(progress.Position, ":")
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
		`, progress.ImdbID, positionAsNumber, c.Locals("UserId"))

	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func ProgressStopped(c *fiber.Ctx) error {
	data := new(struct {
		Completed bool   `form:"completed"`
		ImdbID    string `form:"imdb_id"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	if data.Completed {
		PostMovieNew(c)
	}

	return c.SendStatus(200)
}
