package intake

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// ErrAlreadyWatchlisted is returned when the movie is already on the user's
// watchlist.
var ErrAlreadyWatchlisted = errors.New("movie already on watchlist")

// Watch is one viewing of a movie. See Watch in GLOSSARY.md.
type Watch struct {
	WatchedAt time.Time
	// Friends are the IDs of the users the movie was watched with.
	Friends []string
	// Rating is skipped when zero.
	Rating int
	// Review is skipped when nil. A user keeps their first review of a movie.
	Review *Review
}

type Review struct {
	Content string
	Private bool
}

// LogWatch records that the user watched the movie, adding the movie first
// if needed, and returns the movie's ID. The movie leaves the user's
// watchlist and now playing. Everything is written in one transaction.
func (i *Intake) LogWatch(userID string, ref MovieRef, opts MovieOptions, w Watch) (int, error) {
	p, err := i.resolve(ref)

	if err != nil {
		return 0, err
	}

	tx, err := i.db.Beginx()

	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	movieID, err := i.write(tx, p, opts)

	if err != nil {
		return 0, err
	}

	if err := insertWatch(tx, userID, movieID, w); err != nil {
		return 0, err
	}

	_, err = tx.Exec(`DELETE FROM watchlist WHERE user_id = $1 AND movie_id = $2`, userID, movieID)

	if err != nil {
		return 0, fmt.Errorf("delete watchlist: %w", err)
	}

	_, err = tx.Exec(`DELETE FROM now_playing WHERE user_id = $1 AND movie_id = $2`, userID, movieID)

	if err != nil {
		return 0, fmt.Errorf("delete now playing: %w", err)
	}

	if w.Rating != 0 {
		_, err = tx.Exec(`INSERT INTO rating (user_id, movie_id, rating)
    VALUES ($1, $2, $3)`, userID, movieID, w.Rating)

		if err != nil {
			return 0, fmt.Errorf("insert rating: %w", err)
		}
	}

	if w.Review != nil && w.Review.Content != "" {
		_, err = tx.Exec(`
INSERT INTO review (content, private, user_id, movie_id)
    VALUES ($1, $2, $3, $4)
ON CONFLICT
    DO NOTHING`, w.Review.Content, w.Review.Private, userID, movieID)

		if err != nil {
			return 0, fmt.Errorf("insert review: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	slog.Info("Watch logged", "movie", movieID, "user", userID)

	return movieID, nil
}

// AddToWatchlist puts the movie on the user's watchlist, adding the movie
// first if needed, and returns the movie's ID.
func (i *Intake) AddToWatchlist(userID string, ref MovieRef, opts MovieOptions) (int, error) {
	p, err := i.resolve(ref)

	if err != nil {
		return 0, err
	}

	tx, err := i.db.Beginx()

	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	movieID, err := i.write(tx, p, opts)

	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(`INSERT INTO watchlist (user_id, movie_id)
    VALUES ($1, $2)`, userID, movieID)

	var pqErr *pq.Error

	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return 0, ErrAlreadyWatchlisted
	}

	if err != nil {
		return 0, fmt.Errorf("insert watchlist: %w", err)
	}

	return movieID, tx.Commit()
}

func insertWatch(tx *sqlx.Tx, userID string, movieID int, w Watch) error {
	var seenID int

	err := tx.Get(&seenID, `
INSERT INTO seen (user_id, movie_id, date)
    VALUES ($1, $2, $3)
RETURNING
    id`, userID, movieID, w.WatchedAt)

	if err != nil {
		return fmt.Errorf("insert watch: %w", err)
	}

	if len(w.Friends) == 0 {
		return nil
	}

	_, err = tx.Exec(`
INSERT INTO seen_with (seen_id, other_user_id)
SELECT
    $1,
    UNNEST($2::text[])::int
ON CONFLICT
    DO NOTHING`, seenID, pq.Array(w.Friends))

	if err != nil {
		return fmt.Errorf("insert friends: %w", err)
	}

	return nil
}
