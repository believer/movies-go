package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"

	"github.com/gofiber/fiber/v2"
)

func NowPlaying(c *fiber.Ctx) error {
	var nowPlaying types.Movies

	userId := c.Locals("UserId")

	err := db.Client.Select(&nowPlaying, `
SELECT
    np.position,
    m.id,
    m.title,
    m.runtime,
    m.overview
FROM
    now_playing np
    RIGHT JOIN movie m ON m.imdb_id = np.imdb_id
WHERE
    user_id = $1
			`, userId)

	if err != nil {
		return err
	}

	return utils.Render(c, views.NowPlaying(views.NowPlayingProps{
		Movies: nowPlaying,
	}))
}
