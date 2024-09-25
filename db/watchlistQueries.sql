-- name: watchlist
SELECT
    m.id,
    m.title,
    m.imdb_id,
    w.created_at
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date <= CURRENT_DATE;

-- name: watchlist-unreleased
SELECT
    m.id,
    m.title,
    m.imdb_id,
    m.release_date
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date > CURRENT_DATE
ORDER BY
    m.release_date ASC;

-- name: is-in-watchlist
SELECT
    id
FROM
    watchlist
WHERE
    user_id = $1
    AND movie_id = $2;

