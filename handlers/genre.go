package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetGenre(c *fiber.Ctx) error {
	var movies types.Movies
	var genre db.TableName

	q := db.MakeQueries(c)
	err := q.GetNameByID(&genre, db.GenreTable)

	if err != nil {
		return err
	}

	err = q.GetMovies(&movies, db.GenreTable)

	if err != nil {
		return err
	}

	// When there are no more movies to show, just return 200. Otherwise we
	// would display the "No movies seen" empty state which should only be
	// shown at the start.
	if len(movies) == 0 && q.Page > 1 {
		return c.SendStatus(fiber.StatusOK)
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies for this genre",
		Name:       genre.Name,
		NextPage:   fmt.Sprintf("/genre/%s?page=%d", q.Id, q.Page+1),
		Movies:     movies,
	}))
}

func GetGenreStats(c *fiber.Ctx) error {
	var genres []types.ListItem

	q := db.MakeQueries(c)
	err := q.GetStats(&genres, db.GenreTable)

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  genres,
		Title: "Genre",
		Route: "/genre/stats",
		Root:  "genre",
		Year:  q.Year,
		Years: q.Years,
	}))
}
