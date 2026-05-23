package db

import (
	"believer/movies/types"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ReviewQuerier interface {
	GetReviewByID(id string) (types.Review, error)
	InsertReview(content string, private bool, userID string, movieID int) (types.Review, error)
	DeleteReview(id string) error
	UpdateReview(id string, content string, private bool) (types.Review, error)
}

type ReviewRepository struct {
	db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{db}
}

func (r *ReviewRepository) GetReviewByID(id string) (types.Review, error) {
	var review types.Review
	err := r.db.Get(&review, `
SELECT
    id,
    content,
    private
FROM
    review
WHERE
    id = $1
	`, id)
	return review, err
}

func (r *ReviewRepository) InsertReview(content string, private bool, userID string, movieID int) (types.Review, error) {
	var id int
	err := r.db.Get(&id, `
INSERT INTO review (content, private, user_id, movie_id)
    VALUES ($1, $2, $3, $4)
RETURNING
    id
	`, content, private, userID, movieID)
	if err != nil {
		return types.Review{}, err
	}

	return r.GetReviewByID(fmt.Sprintf("%d", id))
}

func (r *ReviewRepository) UpdateReview(id string, content string, private bool) (types.Review, error) {
	_, err := r.db.Exec(`
UPDATE
    review
SET
    content = $1,
    private = $2
WHERE
    id = $3
	`, content, private, id)
	if err != nil {
		return types.Review{}, err
	}

	return r.GetReviewByID(id)
}

func (r *ReviewRepository) DeleteReview(id string) error {
	_, err := r.db.Exec(`
DELETE FROM review
WHERE id = $1
	`, id)
	return err
}
