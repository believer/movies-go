package db

import (
	"believer/movies/types"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type ReviewQuerier interface {
	GetReviewByID(id string, userID string) (types.Review, error)
	InsertReview(content string, private bool, userID string, movieID int) (types.Review, error)
	DeleteReview(id string, userID string) error
	UpdateReview(id string, userID string, content string, private bool) (types.Review, error)
}

type ReviewRepository struct {
	db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{db}
}

// GetReviewByID returns the user's Review, or sql.ErrNoRows when the user
// has no Review with the ID.
func (r *ReviewRepository) GetReviewByID(id string, userID string) (types.Review, error) {
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
    AND user_id = $2
	`, id, userID)
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

	return r.GetReviewByID(fmt.Sprintf("%d", id), userID)
}

// UpdateReview updates the user's Review, or returns sql.ErrNoRows when the
// user has no Review with the ID.
func (r *ReviewRepository) UpdateReview(id string, userID string, content string, private bool) (types.Review, error) {
	res, err := r.db.Exec(`
UPDATE
    review
SET
    content = $1,
    private = $2
WHERE
    id = $3
    AND user_id = $4
	`, content, private, id, userID)
	if err != nil {
		return types.Review{}, err
	}

	if err := requireRow(res); err != nil {
		return types.Review{}, err
	}

	return r.GetReviewByID(id, userID)
}

// DeleteReview deletes the user's Review, or returns sql.ErrNoRows when the
// user has no Review with the ID.
func (r *ReviewRepository) DeleteReview(id string, userID string) error {
	res, err := r.db.Exec(`
DELETE FROM review
WHERE id = $1
    AND user_id = $2
	`, id, userID)
	if err != nil {
		return err
	}

	return requireRow(res)
}
