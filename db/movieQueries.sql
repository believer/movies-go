-- name: movie-by-id
SELECT
    m.id,
    m.title,
    m.release_date,
    m.runtime,
    m.imdb_id,
    m.overview,
    m.original_title,
    m.tagline,
    se.name AS "series",
    se.id AS "series_id",
    ms.number_in_series,
    r.rating,
    COALESCE(ARRAY_TO_JSON(ARRAY_AGG(DISTINCT jsonb_build_object('name', g.name, 'id', g.id)) FILTER (WHERE g.name IS NOT NULL)), '[]') AS genres,
    COALESCE(ARRAY_TO_JSON(ARRAY_AGG(DISTINCT jsonb_build_object('name', l.english_name, 'id', l.id)) FILTER (WHERE l.english_name IS NOT NULL)), '[]') AS languages
FROM
    movie AS m
    LEFT JOIN movie_genre AS mg ON mg.movie_id = m.id
    LEFT JOIN genre AS g ON g.id = mg.genre_id
    LEFT JOIN rating AS r ON r.movie_id = m.id
        AND r.user_id = $2
    LEFT JOIN movie_series AS ms ON ms.movie_id = m.id
    LEFT JOIN series AS se ON se.id = ms.series_id
    LEFT JOIN movie_language AS ml ON ml.movie_id = m.id
    LEFT JOIN "language" AS l ON l.id = ml.language_id
WHERE
    m.id = $1
GROUP BY
    1,
    r.rating,
    se.id,
    ms.number_in_series;

-- name: review-by-movie-id
SELECT
    id,
    content,
    private
FROM
    review
WHERE
    movie_id = $1
    AND user_id = $2;

-- name: insert-movie
INSERT INTO movie (title, runtime, release_date, imdb_id, overview, poster, tagline, wilhelm)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (imdb_id)
    DO UPDATE SET
        title = $1
    RETURNING
        id;

-- name: insert-review
INSERT INTO review (content, private, user_id, movie_id)
    VALUES ($1, $2, $3, $4)
ON CONFLICT
    DO NOTHING;

