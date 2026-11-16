package db

import (
	"believer/movies/views"
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
	// user2Watch is user 2's Watch, seen with user 3
	user2Watch string
}

// setupWatch gives user 2 a Watch of a movie, seen with user 3, for tests
// that act on it as user 1. User 1 has a Watch of the movie too, so a rule
// like "any Watch the user has" doesn't pass.
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
	ts.watch(t, 1, movieID, "2026-01-03 20:00")

	return testWatch{ts, NewMovieRepository(ts.db), strconv.Itoa(seenID)}
}

// assertUntouched checks that user 2's Watch keeps its date and seen-with.
func (tw testWatch) assertUntouched(t *testing.T) {
	t.Helper()

	watch, err := tw.repo.GetSeenMovie(tw.user2Watch, "2")
	require.NoError(t, err)
	assert.Equal(t, "2026-01-02T21:00", watch.Date)
	assert.Equal(t, []string{"3"}, []string(watch.SeenWith))
}

func TestGetSeenMovie(t *testing.T) {
	tw := setupWatch(t)

	t.Run("returns the user's Watch", func(t *testing.T) {
		watch, err := tw.repo.GetSeenMovie(tw.user2Watch, "2")

		require.NoError(t, err)
		assert.Equal(t, "2026-01-02T21:00", watch.Date)
	})

	t.Run("doesn't return another user's Watch", func(t *testing.T) {
		_, err := tw.repo.GetSeenMovie(tw.user2Watch, "1")

		assert.ErrorIs(t, err, sql.ErrNoRows)
	})
}

func TestUpdateSeenMovie(t *testing.T) {
	t.Run("doesn't update another user's Watch", func(t *testing.T) {
		tw := setupWatch(t)

		err := tw.repo.UpdateSeenMovie(tw.user2Watch, "1", time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC), []string{"1"})

		assert.ErrorIs(t, err, sql.ErrNoRows)
		tw.assertUntouched(t)
	})

	t.Run("updates the user's Watch", func(t *testing.T) {
		tw := setupWatch(t)

		err := tw.repo.UpdateSeenMovie(tw.user2Watch, "2", time.Date(2026, 5, 6, 18, 0, 0, 0, time.UTC), []string{"1"})
		require.NoError(t, err)

		watch, err := tw.repo.GetSeenMovie(tw.user2Watch, "2")
		require.NoError(t, err)
		assert.Equal(t, "2026-05-06T20:00", watch.Date)
		assert.Equal(t, []string{"1"}, []string(watch.SeenWith))
	})
}

// person adds a person and returns their ID.
func (ts testStats) person(t *testing.T, name string, popularity float64) int {
	t.Helper()

	var id int

	if err := ts.db.Get(&id, `INSERT INTO person (name, original_id, popularity) SELECT $1, COALESCE(MAX(original_id), 0) + 1, $2 FROM person RETURNING id`, name, popularity); err != nil {
		t.Fatal(err)
	}

	return id
}

func (ts testStats) credit(t *testing.T, movieID int, personID int, job string) {
	t.Helper()
	ts.exec(t, `INSERT INTO movie_person (movie_id, person_id, job) VALUES ($1, $2, $3)`, movieID, personID, job)
}

func castNames(people []views.CastAndCrewDTO) []string {
	names := make([]string, len(people))

	for i, p := range people {
		names[i] = p.Name
	}

	return names
}

func TestCast(t *testing.T) {
	t.Run("orders people by their movies in any job, then popularity", func(t *testing.T) {
		ts := setupStats(t)
		heat := ts.movie(t, "Heat")
		other := ts.movie(t, "Other")
		another := ts.movie(t, "Another")

		// Director credits count, so counting only cast credits would put
		// Cast twice first
		director := ts.person(t, "Directs too", 1)
		ts.credit(t, heat, director, "cast")
		ts.credit(t, other, director, "director")
		ts.credit(t, another, director, "director")

		twice := ts.person(t, "Cast twice", 5)
		ts.credit(t, heat, twice, "cast")
		ts.credit(t, other, twice, "cast")

		popular := ts.person(t, "Popular", 9)
		ts.credit(t, heat, popular, "cast")

		cast, _, err := NewMovieRepository(ts.db).Cast(strconv.Itoa(heat))

		require.NoError(t, err)
		require.Len(t, cast, 1)
		assert.Equal(t, "Cast", cast[0].Job)
		assert.Equal(t, []string{"Directs too", "Cast twice", "Popular"}, castNames(cast[0].People))
	})

	t.Run("breaks ties on movies, popularity and name by ID", func(t *testing.T) {
		ts := setupStats(t)
		heat := ts.movie(t, "Heat")
		first := ts.person(t, "Namesake", 1)
		second := ts.person(t, "Namesake", 1)

		// Credited in reverse ID order, so the order of the rows doesn't
		// pass for the tiebreak
		ts.credit(t, heat, second, "cast")
		ts.credit(t, heat, first, "cast")

		cast, _, err := NewMovieRepository(ts.db).Cast(strconv.Itoa(heat))

		require.NoError(t, err)
		require.Len(t, cast, 1)
		assert.Equal(t, []int32{int32(first), int32(second)}, []int32{cast[0].People[0].ID, cast[0].People[1].ID})
	})
}

func TestDeleteSeenMovie(t *testing.T) {
	deleteAs := func(t *testing.T, tw testWatch, userID string) error {
		t.Helper()

		tx, err := tw.repo.Begin()
		require.NoError(t, err)
		defer tx.Rollback()

		if err := tw.repo.DeleteSeenMovie(tx, tw.user2Watch, userID); err != nil {
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

		_, err := tw.repo.GetSeenMovie(tw.user2Watch, "2")
		assert.ErrorIs(t, err, sql.ErrNoRows)

		var seenWith int
		require.NoError(t, tw.db.Get(&seenWith, `SELECT COUNT(*) FROM seen_with WHERE seen_id = $1`, tw.user2Watch))
		assert.Zero(t, seenWith)
	})
}
