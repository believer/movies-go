package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

func GetLanguage(c *fiber.Ctx) error {
	var language db.TableName
	var movies types.Movies

	q := db.MakeQueries(c)
	err := q.GetNameByID(&language, db.LanguageTable)

	if err != nil {
		return err
	}

	err = q.GetMovies(&movies, db.LanguageTable)

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
		EmptyState: "No movies for this language",
		Name:       language.Name,
		NextPage:   fmt.Sprintf("/language/%s?page=%d", q.Id, q.Page+1),
		Movies:     movies,
	}))
}

func GetLanguageStats(c *fiber.Ctx) error {
	var languages []types.ListItem

	q := db.MakeQueries(c)
	err := q.GetStats(&languages, db.LanguageTable)

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  languages,
		Title: "Language",
		Root:  "language",
		Route: "/language/stats",
		Year:  q.Year,
		Years: q.Years,
	}))
}
