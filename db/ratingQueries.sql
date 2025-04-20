-- name: movies-by-rating
SELECT DISTINCT
    (m.id),
    m.title,
    m.release_date,
    m.imdb_id,
    (s.id IS NOT NULL) AS "seen"
FROM
    rating r
    INNER JOIN movie m ON m.id = r.movie_id
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
    r.rating = $1
    AND r.user_id = $2
ORDER BY
    m.release_date DESC OFFSET $3
LIMIT 50;

