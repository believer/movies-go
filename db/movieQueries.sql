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

