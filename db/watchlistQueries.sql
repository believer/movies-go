-- name: watchlist
SELECT
    m.id,
    m.title,
    m.imdb_id,
    m.release_date,
    w.created_at
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date <= CURRENT_DATE
ORDER BY
    CASE WHEN $2 = 'Release date' THEN
        m.release_date
    ELSE
        w.created_at
    END ASC;

-- name: watchlist-unreleased
SELECT
    m.id,
    m.title,
    m.imdb_id,
    m.release_date,
    w.created_at
FROM
    watchlist w
    INNER JOIN movie m ON m.id = w.movie_id
WHERE
    user_id = $1
    AND m.release_date > CURRENT_DATE
ORDER BY
    CASE WHEN $2 = 'Date added' THEN
        w.created_at
    ELSE
        m.release_date
    END ASC;

-- name: is-in-watchlist
SELECT
    id
FROM
    watchlist
WHERE
    user_id = $1
    AND movie_id = $2;

