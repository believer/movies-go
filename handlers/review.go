package handlers

import (
	"believer/movies/components/review"
	"believer/movies/db"
	"believer/movies/types"

	"github.com/gofiber/fiber/v2"
)

type ReviewHandler struct {
	repo db.ReviewQuerier
}

func NewReviewHandler(repo db.ReviewQuerier) *ReviewHandler {
	return &ReviewHandler{repo}
}

func (h *ReviewHandler) AddMovieReview(c *fiber.Ctx) error {
	req := NewRequest(c)
	isAuth := req.IsAuthenticated()
	movieID := req.Query("movieId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	return Render(c, review.AddReview(movieID))
}

func (h *ReviewHandler) InsertMovieReview(c *fiber.Ctx) error {
	req := NewRequest(c)
	isAuth := req.IsAuthenticated()
	movieID := req.QueryInt("movieId")
	userID := req.UserID()

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

	InvalidateStatsCache(userID)

	return Render(c, review.Review(reviewData, movieID))
}

func (h *ReviewHandler) EditMovieReview(c *fiber.Ctx) error {
	req := NewRequest(c)
	isAuth := req.IsAuthenticated()
	movieID := req.QueryInt("movieId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	id := req.Params("id")
	reviewData, err := h.repo.GetReviewByID(id)
	if err != nil {
		return err
	}

	return Render(c, review.EditReview(reviewData, movieID))
}

func (h *ReviewHandler) UpdateMovieReview(c *fiber.Ctx) error {
	req := NewRequest(c)
	movieID := req.QueryInt("movieId")
	id := req.Params("id")
	isAuth := req.IsAuthenticated()

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

	if userID := req.UserID(); userID != "" {
		InvalidateStatsCache(userID)
	}

	return Render(c, review.ReviewContent(reviewData, movieID))
}

func (h *ReviewHandler) DeleteMovieReview(c *fiber.Ctx) error {
	req := NewRequest(c)
	id := req.Params("id")
	isAuth := req.IsAuthenticated()
	movieID := req.QueryInt("movieId")

	if !isAuth {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	err := h.repo.DeleteReview(id)
	if err != nil {
		return err
	}

	if userID := req.UserID(); userID != "" {
		InvalidateStatsCache(userID)
	}

	return Render(c, review.Review(types.Review{}, movieID))
}
