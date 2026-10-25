package intake

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedUsers(t *testing.T, db *sqlx.DB) {
	t.Helper()

	_, err := db.Exec(`
INSERT INTO "user" (id, username, password_hash, name)
    VALUES (1, 'me', '', 'Me'), (2, 'friend', '', 'Friend')`)
	require.NoError(t, err)
}

func TestLogWatch(t *testing.T) {
	watchedAt := time.Date(2026, 10, 20, 19, 45, 0, 0, time.UTC)

	t.Run("adds the movie and records the watch", func(t *testing.T) {
		in, _, db := setup(t)
		seedUsers(t, db)

		id, err := in.LogWatch("1", ByExternalID("tt1375666"), MovieOptions{}, Watch{
			WatchedAt: watchedAt,
			Friends:   []string{"2"},
			Rating:    9,
			Review:    &Review{Content: "Dreams within dreams", Private: true},
		})
		require.NoError(t, err)

		var date time.Time
		require.NoError(t, db.Get(&date, `SELECT date FROM seen WHERE movie_id = $1 AND user_id = 1`, id))
		assert.True(t, watchedAt.Equal(date), "watched at %s, stored %s", watchedAt, date)

		assert.Equal(t, 1, count(t, db, `SELECT count(*) FROM seen_with sw INNER JOIN seen s ON s.id = sw.seen_id WHERE s.movie_id = $1 AND sw.other_user_id = 2`, id))
		assert.Equal(t, 9, count(t, db, `SELECT rating FROM rating WHERE movie_id = $1 AND user_id = 1`, id))
		assert.Equal(t, 1, count(t, db, `SELECT count(*) FROM review WHERE movie_id = $1 AND user_id = 1 AND private`, id))
	})

	t.Run("removes the movie from watchlist and now playing", func(t *testing.T) {
		in, _, db := setup(t)
		seedUsers(t, db)

		id, err := in.AddToWatchlist("1", ByExternalID("tt1375666"), MovieOptions{})
		require.NoError(t, err)

		_, err = db.Exec(`INSERT INTO now_playing (movie_id, position, user_id) VALUES ($1, 42, 1)`, id)
		require.NoError(t, err)

		_, err = in.LogWatch("1", ByID(id), MovieOptions{}, Watch{WatchedAt: watchedAt})
		require.NoError(t, err)

		assert.Equal(t, 0, count(t, db, `SELECT count(*) FROM watchlist WHERE movie_id = $1`, id))
		assert.Equal(t, 0, count(t, db, `SELECT count(*) FROM now_playing WHERE movie_id = $1`, id))
	})

	t.Run("rolls back the new movie when the watch fails", func(t *testing.T) {
		in, _, db := setup(t)
		seedUsers(t, db)

		// User 999 doesn't exist, so inserting the friend fails
		_, err := in.LogWatch("1", ByExternalID("tt1375666"), MovieOptions{}, Watch{
			WatchedAt: watchedAt,
			Friends:   []string{"999"},
		})
		assert.Error(t, err)

		assert.Equal(t, 0, count(t, db, `SELECT count(*) FROM movie`))
		assert.Equal(t, 0, count(t, db, `SELECT count(*) FROM seen`))
	})

	t.Run("records each watch of the same movie", func(t *testing.T) {
		in, _, db := setup(t)
		seedUsers(t, db)

		id, err := in.LogWatch("1", ByExternalID("tt1375666"), MovieOptions{}, Watch{WatchedAt: watchedAt})
		require.NoError(t, err)

		_, err = in.LogWatch("1", ByID(id), MovieOptions{}, Watch{WatchedAt: watchedAt.AddDate(1, 0, 0)})
		require.NoError(t, err)

		assert.Equal(t, 2, count(t, db, `SELECT count(*) FROM seen WHERE movie_id = $1`, id))
	})
}

func TestAddToWatchlist(t *testing.T) {
	t.Run("adds the movie without a watch", func(t *testing.T) {
		in, _, db := setup(t)
		seedUsers(t, db)

		id, err := in.AddToWatchlist("1", ByExternalID("tt1375666"), MovieOptions{})
		require.NoError(t, err)

		assert.Equal(t, 1, count(t, db, `SELECT count(*) FROM watchlist WHERE movie_id = $1 AND user_id = 1`, id))
		assert.Equal(t, 0, count(t, db, `SELECT count(*) FROM seen`))
	})

	t.Run("returns ErrAlreadyWatchlisted", func(t *testing.T) {
		in, _, db := setup(t)
		seedUsers(t, db)

		_, err := in.AddToWatchlist("1", ByExternalID("tt1375666"), MovieOptions{})
		require.NoError(t, err)

		_, err = in.AddToWatchlist("1", ByExternalID("tt1375666"), MovieOptions{})
		assert.ErrorIs(t, err, ErrAlreadyWatchlisted)
	})
}
