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
    COALESCE(ARRAY_TO_JSON(ARRAY (
                SELECT
                    jsonb_build_object('id', id, 'name', name)
                FROM ( SELECT DISTINCT ON (pc.id)
                    pc.id, pc.name FROM production_company pc
                    JOIN movie_company mc2 ON mc2.company_id = pc.id
                    WHERE
                        mc2.movie_id = m.id ORDER BY pc.id, pc.name) AS uniq_pc ORDER BY name ASC)), '[]') AS production_companies,
    COALESCE(ARRAY_TO_JSON(ARRAY (
                SELECT
                    jsonb_build_object('id', id, 'name', name)
                FROM ( SELECT DISTINCT ON (pc.id)
                        pc.id, pc.name
                    FROM production_country pc
                    JOIN movie_country mc2 ON mc2.country_id = pc.id
                    WHERE
                        mc2.movie_id = m.id ORDER BY pc.id, pc.name) AS uniq_pc ORDER BY name ASC)), '[]') AS production_countries,
    r.created_at at time zone 'UTC' at time zone 'Europe/Stockholm' AS "rated_at",
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
    r.id,
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

-- name: review-by-id
SELECT
    id,
    content,
    private
FROM
    review
WHERE
    id = $1;

-- name: insert-movie
INSERT INTO movie (title, runtime, release_date, imdb_id, overview, poster, tagline, tmdb_id, wilhelm)
    VALUES ($1, $2, NULLIF ($3, '')::date, $4, $5, $6, $7, $8, $9)
ON CONFLICT (imdb_id)
    DO UPDATE SET
        title = excluded.title,
        runtime = excluded.runtime,
        release_date = excluded.release_date,
        imdb_id = excluded.imdb_id,
        overview = excluded.overview,
        poster = excluded.poster,
        tagline = excluded.tagline,
        tmdb_id = excluded.tmdb_id
    RETURNING
        id;

-- name: insert-review
INSERT INTO review (content, private, user_id, movie_id)
    VALUES ($1, $2, $3, $4)
ON CONFLICT
    DO NOTHING;

-- name: movie-awards
SELECT
    name AS category,
    year,
    COALESCE(JSONB_AGG(
            CASE WHEN person IS NOT NULL
                AND person_id IS NOT NULL THEN
                JSONB_BUILD_OBJECT('name', person, 'id', person_id)
            WHEN person IS NOT NULL THEN
                JSONB_BUILD_OBJECT('name', person)
            ELSE
                JSONB_BUILD_OBJECT('name', 'N/A')
            END) FILTER (WHERE person IS NOT NULL
            OR person_id IS NOT NULL), '[]'::jsonb) AS nominees,
    winner,
    detail
FROM
    award
WHERE
    imdb_id = $1
GROUP BY
    name,
    year,
    winner,
    detail
ORDER BY
    winner DESC,
    category ASC;

-- name: others-ratings
SELECT
    (
        SELECT
            count(DISTINCT user_id)
        FROM
            seen
        WHERE
            movie_id = $1) AS seen_count,
    (
        SELECT
            COALESCE(AVG(r.latest_rating), 0)
        FROM ( SELECT DISTINCT ON (user_id)
                user_id,
                rating AS latest_rating
            FROM
                rating
            WHERE
                movie_id = $1
            ORDER BY
                user_id,
                created_at DESC) r) AS avg_rating;

