package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type AwardsHandler struct {
	repo db.AwardsQuerier
}

func NewAwardsHandler(repo db.AwardsQuerier) *AwardsHandler {
	return &AwardsHandler{repo}
}

func (h *AwardsHandler) GetMoviesByNumberOfAwards(c *fiber.Ctx) error {
	userID := c.Locals("UserId").(string)
	numberOfAwards, err := c.ParamsInt("awards")
	if err != nil {
		return err
	}

	includeNominations := c.QueryBool("nominations")
	awardType := c.Query("type")

	cfg, err := awardConfigFromQuery(c)

	if err != nil {
		return err
	}

	var movies types.Movies
	var name string

	if includeNominations {
		name = fmt.Sprintf(cfg.NominationName, numberOfAwards)
		movies, err = h.repo.GetByNominations(userID, numberOfAwards, awardType)
	} else {
		name = fmt.Sprintf(cfg.WinName, numberOfAwards)
		movies, err = h.repo.GetByWins(userID, numberOfAwards, awardType)
	}

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: cfg.EmptyState,
		Name:       name,
		Movies:     movies,
	}))
}

func (h *AwardsHandler) GetAwardsByYear(c *fiber.Ctx) error {
	year := c.Params("year")
	sort := c.Query("sort", "Movie")
	awardType := c.Query("type")

	_, err := awardConfigFromQuery(c)

	if err != nil {
		return err
	}

	// Only valid types
	validSorts := map[string]bool{"Movie": true, "Category": true}
	if !validSorts[sort] {
		err := fiber.ErrBadRequest
		err.Message = "Invalid sort value"
		return err
	}

	switch sort {
	case "Movie":
		awards, err := h.repo.GetGroupedByMovie(year, awardType)

		if err != nil {
			return err
		}

		return utils.Render(c, views.AwardsPage(views.AwardsPageProps{
			GroupedAwards: awards,
			Sort:          sort,
			Type:          awardType,
			Year:          year,
		}))
	case "Category":
		awards, err := h.repo.GetGroupedByCategory(year, awardType)

		if err != nil {
			return err
		}

		return utils.Render(c, views.AwardsCategory(views.AwardsCategoryProps{
			Awards: awards,
			Sort:   sort,
			Type:   awardType,
			Year:   year,
		}))
	}

	return utils.Render(c, views.NotFound())
}

func awardConfigFromQuery(c *fiber.Ctx) (types.AwardConfig, error) {
	awardType := c.Query("type")

	if awardType == "" {
		err := fiber.ErrBadRequest
		err.Message = "Missing awardType"
		return types.AwardConfig{}, err
	}

	cfg, ok := types.GetAwardConfig(awardType)
	if !ok {
		err := fiber.ErrBadRequest
		err.Message = "Incompatible awardType"
		return types.AwardConfig{}, err
	}

	return cfg, nil
}
