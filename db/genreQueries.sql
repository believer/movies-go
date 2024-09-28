-- name: genres-by-id
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    m.imdb_id,
    (s.id IS NOT NULL) AS "seen"
FROM
    movie_genre mg
    INNER JOIN movie m ON m.id = mg.movie_id
    LEFT JOIN ( SELECT DISTINCT ON (movie_id)
            movie_id,
            id
        FROM
            public.seen
        WHERE
            user_id = $2
        ORDER BY
            movie_id,
            id) AS s ON m.id = s.movie_id
WHERE
    mg.genre_id = $1
ORDER BY
    m.release_date DESC OFFSET $3
LIMIT 50;

-- name: genre-by-id
SELECT
    name
FROM
    genre
WHERE
    id = $1;

-- name: movies-in-genre
SELECT
    count(*)
FROM
    movie_genre
WHERE
    id = $1;

