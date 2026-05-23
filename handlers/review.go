package handlers

import (
	"believer/movies/components/review"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"

	"github.com/gofiber/fiber/v2"
)

type ReviewHandler struct {
	repo db.ReviewQuerier
}

func NewReviewHandler(repo db.ReviewQuerier) *ReviewHandler {
	return &ReviewHandler{repo}
}

func (h *ReviewHandler) AddMovieReview(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieID := c.Query("movieId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return utils.Render(c, review.AddReview(movieID))
}

func (h *ReviewHandler) InsertMovieReview(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieID := c.QueryInt("movieId")
	userID, ok := c.Locals("UserId").(string)
	if !ok {
		userID = ""
	}

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		IsPrivate bool   `form:"review_private"`
		Review    string `form:"review"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	reviewData, err := h.repo.InsertReview(data.Review, data.IsPrivate, userID, movieID)
	if err != nil {
		return err
	}

	return utils.Render(c, review.Review(reviewData, movieID))
}

func (h *ReviewHandler) EditMovieReview(c *fiber.Ctx) error {
	isAuth := utils.IsAuthenticated(c)
	movieID := c.QueryInt("movieId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	id := c.Params("id")
	reviewData, err := h.repo.GetReviewByID(id)
	if err != nil {
		return err
	}

	return utils.Render(c, review.EditReview(reviewData, movieID))
}

func (h *ReviewHandler) UpdateMovieReview(c *fiber.Ctx) error {
	movieID := c.QueryInt("movieId")
	id := c.Params("id")
	isAuth := utils.IsAuthenticated(c)

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	data := new(struct {
		Review          string `form:"review"`
		IsPrivateReview bool   `form:"review_private"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	reviewData, err := h.repo.UpdateReview(id, data.Review, data.IsPrivateReview)
	if err != nil {
		return err
	}

	return utils.Render(c, review.ReviewContent(reviewData, movieID))
}

func (h *ReviewHandler) DeleteMovieReview(c *fiber.Ctx) error {
	id := c.Params("id")
	isAuth := utils.IsAuthenticated(c)
	movieID := c.QueryInt("movieId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	err := h.repo.DeleteReview(id)
	if err != nil {
		return err
	}

	return utils.Render(c, review.Review(types.Review{}, movieID))
}
