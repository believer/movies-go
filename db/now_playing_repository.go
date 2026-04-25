package db

import (
	"believer/movies/types"

	"github.com/jmoiron/sqlx"
)

// Repo
// =====================================================

type NowPlayingQuerier interface {
	GetNowPlaying(userID string) (types.Movies, error)
}

type NowPlayingRepository struct {
	db *sqlx.DB
}

func NewNowPlayingRepository(db *sqlx.DB) *NowPlayingRepository {
	return &NowPlayingRepository{db}
}

func (r *NowPlayingRepository) GetNowPlaying(userID string) (types.Movies, error) {
	var movies types.Movies
	err := r.db.Select(&movies, nowPlayingQuery, userID)
	return movies, err
}

// Queries
// =====================================================

const nowPlayingQuery = `
SELECT
    np.position,
    m.id,
    m.title,
    m.runtime,
    m.overview,
    (
        CASE WHEN m.runtime != 0 THEN
            np."position" / m.runtime
        ELSE
            0
        END) AS percent
FROM
    now_playing np
    RIGHT JOIN movie m ON m.imdb_id = np.imdb_id
WHERE
    user_id = $1
ORDER BY
    percent DESC
`
