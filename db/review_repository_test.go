package db

import (
	"database/sql"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testReview struct {
	testStats
	repo *ReviewRepository
	// user2Review is user 2's private Review
	user2Review string
}

// setupReview gives user 2 a private Review of a movie, for tests that act
// on it as user 1. User 1 has a Review of the movie too, so a rule like "any
// Review of the movie" doesn't pass.
func setupReview(t *testing.T) testReview {
	t.Helper()

	ts := setupStats(t)
	movieID := ts.movie(t, "Heat")

	var reviewID int

	if err := ts.db.Get(&reviewID, `INSERT INTO review (user_id, movie_id, content, private) VALUES (2, $1, 'Secret', true) RETURNING id`, movieID); err != nil {
		t.Fatal(err)
	}

	ts.exec(t, `INSERT INTO review (user_id, movie_id, content, private) VALUES (1, $1, 'Mine', false)`, movieID)

	return testReview{ts, NewReviewRepository(ts.db), strconv.Itoa(reviewID)}
}

// assertUntouched checks that user 2's Review keeps its content and privacy.
func (tr testReview) assertUntouched(t *testing.T) {
	t.Helper()

	review, err := tr.repo.GetReviewByID(tr.user2Review, "2")
	require.NoError(t, err)
	assert.Equal(t, "Secret", review.Content)
	assert.True(t, review.Private)
}

func TestGetReviewByID(t *testing.T) {
	tr := setupReview(t)

	t.Run("returns the user's Review", func(t *testing.T) {
		review, err := tr.repo.GetReviewByID(tr.user2Review, "2")

		require.NoError(t, err)
		assert.Equal(t, "Secret", review.Content)
	})

	t.Run("doesn't return another user's Review", func(t *testing.T) {
		_, err := tr.repo.GetReviewByID(tr.user2Review, "1")

		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestUpdateReview(t *testing.T) {
	t.Run("doesn't update another user's Review", func(t *testing.T) {
		tr := setupReview(t)

		_, err := tr.repo.UpdateReview(tr.user2Review, "1", "Changed", false)

		assert.ErrorIs(t, err, sql.ErrNoRows)
		tr.assertUntouched(t)
	})

	t.Run("updates the user's Review", func(t *testing.T) {
		tr := setupReview(t)

		review, err := tr.repo.UpdateReview(tr.user2Review, "2", "Changed", false)

		require.NoError(t, err)
		assert.Equal(t, "Changed", review.Content)
		assert.False(t, review.Private)
	})
}

func TestDeleteReview(t *testing.T) {
	t.Run("doesn't delete another user's Review", func(t *testing.T) {
		tr := setupReview(t)

		err := tr.repo.DeleteReview(tr.user2Review, "1")

		assert.ErrorIs(t, err, sql.ErrNoRows)
		tr.assertUntouched(t)
	})

	t.Run("deletes the user's Review", func(t *testing.T) {
		tr := setupReview(t)

		err := tr.repo.DeleteReview(tr.user2Review, "2")
		require.NoError(t, err)

		_, err = tr.repo.GetReviewByID(tr.user2Review, "2")
		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}
