package db

import (
	"database/sql"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testWatch struct {
	testStats
	repo *MovieRepository
	// otherWatch is user 2's Watch, seen with user 3
	otherWatch string
}

// setupWatch gives user 2 a Watch of a movie, seen with user 3, for tests
// that act on it as user 1.
func setupWatch(t *testing.T) testWatch {
	t.Helper()

	ts := setupStats(t)
	ts.seedUser(t, 3)
	movieID := ts.movie(t, "Heat")

	var seenID int

	if err := ts.db.Get(&seenID, `INSERT INTO seen (user_id, movie_id, date) VALUES (2, $1, '2026-01-02 20:00') RETURNING id`, movieID); err != nil {
		t.Fatal(err)
	}

	ts.exec(t, `INSERT INTO seen_with (seen_id, other_user_id) VALUES ($1, 3)`, seenID)

	return testWatch{ts, NewMovieRepository(ts.db), strconv.Itoa(seenID)}
}

// assertUntouched checks that user 2's Watch keeps its date and seen-with.
func (tw testWatch) assertUntouched(t *testing.T) {
	t.Helper()

	watch, err := tw.repo.GetSeenMovie(tw.otherWatch, "2")
	require.NoError(t, err)
	assert.Equal(t, "2026-01-02T21:00", watch.Date)
	assert.Equal(t, []string{"3"}, []string(watch.SeenWith))
}

func TestGetSeenMovie(t *testing.T) {
	tw := setupWatch(t)

	t.Run("returns the user's Watch", func(t *testing.T) {
		watch, err := tw.repo.GetSeenMovie(tw.otherWatch, "2")

		require.NoError(t, err)
		assert.Equal(t, "2026-01-02T21:00", watch.Date)
	})

	t.Run("doesn't return another user's Watch", func(t *testing.T) {
		_, err := tw.repo.GetSeenMovie(tw.otherWatch, "1")

		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestUpdateSeenMovie(t *testing.T) {
	t.Run("doesn't update another user's Watch", func(t *testing.T) {
		tw := setupWatch(t)

		err := tw.repo.UpdateSeenMovie(tw.otherWatch, "1", time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC), []string{"1"})

		assert.ErrorIs(t, err, sql.ErrNoRows)
		tw.assertUntouched(t)
	})

	t.Run("updates the user's Watch", func(t *testing.T) {
		tw := setupWatch(t)

		err := tw.repo.UpdateSeenMovie(tw.otherWatch, "2", time.Date(2026, 5, 6, 18, 0, 0, 0, time.UTC), []string{"1"})
		require.NoError(t, err)

		watch, err := tw.repo.GetSeenMovie(tw.otherWatch, "2")
		require.NoError(t, err)
		assert.Equal(t, "2026-05-06T20:00", watch.Date)
		assert.Equal(t, []string{"1"}, []string(watch.SeenWith))
	})
}

func TestDeleteSeenMovie(t *testing.T) {
	deleteAs := func(t *testing.T, tw testWatch, userID string) error {
		t.Helper()

		tx, err := tw.repo.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		if err := tw.repo.DeleteSeenMovie(tx, tw.otherWatch, userID); err != nil {
			return err
		}

		return tx.Commit()
	}

	t.Run("doesn't delete another user's Watch", func(t *testing.T) {
		tw := setupWatch(t)

		assert.ErrorIs(t, deleteAs(t, tw, "1"), sql.ErrNoRows)
		tw.assertUntouched(t)
	})

	t.Run("deletes the user's Watch", func(t *testing.T) {
		tw := setupWatch(t)

		require.NoError(t, deleteAs(t, tw, "2"))

		_, err := tw.repo.GetSeenMovie(tw.otherWatch, "2")
		assert.ErrorIs(t, err, sql.ErrNoRows)

		var seenWith int
		require.NoError(t, tw.db.Get(&seenWith, `SELECT COUNT(*) FROM seen_with WHERE seen_id = $1`, tw.otherWatch))
		assert.Zero(t, seenWith)
	})
}
