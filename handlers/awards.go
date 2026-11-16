package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
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
	req := NewRequest(c)
	userID := req.UserID()
	numberOfAwards, err := req.ParamsInt("awards")
	if err != nil {
		return err
	}

	includeNominations := req.QueryBool("nominations")
	awardType := req.Query("type")

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

	return Render(c, views.ListView(views.ListViewProps{
		EmptyState: cfg.EmptyState,
		Name:       name,
		Movies:     movies,
	}))
}

func (h *AwardsHandler) GetAwardsByYear(c *fiber.Ctx) error {
	req := NewRequest(c)
	year := req.Params("year")
	sort := req.QueryDefault("sort", "Movie")
	awardType := req.Query("type")

	_, err := awardConfigFromQuery(c)

	if err != nil {
		return err
	}

	// Only valid types
	validSorts := map[string]bool{"Movie": true, "Category": true}
	if !validSorts[sort] {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid sort value")
	}

	switch sort {
	case "Movie":
		awards, err := h.repo.GetGroupedByMovie(year, awardType)

		if err != nil {
			return err
		}

		return Render(c, views.AwardsPage(views.AwardsPageProps{
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

		return Render(c, views.AwardsCategory(views.AwardsCategoryProps{
			Awards: awards,
			Sort:   sort,
			Type:   awardType,
			Year:   year,
		}))
	}

	return Render(c, views.NotFound())
}

func awardConfigFromQuery(c *fiber.Ctx) (types.AwardConfig, error) {
	req := NewRequest(c)
	awardType := req.Query("type")

	if awardType == "" {
		return types.AwardConfig{}, fiber.NewError(fiber.StatusBadRequest, "Missing awardType")
	}

	cfg, ok := types.GetAwardConfig(awardType)
	if !ok {
		return types.AwardConfig{}, fiber.NewError(fiber.StatusBadRequest, "Incompatible awardType")
	}

	return cfg, nil
}
