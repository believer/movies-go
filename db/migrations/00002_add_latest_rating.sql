-- +goose Up
-- A user can rate a movie again, and the latest Rating counts (GLOSSARY.md).
-- One row per user and movie: the latest by created_at, then by id.
CREATE VIEW latest_rating AS
SELECT DISTINCT ON (user_id, movie_id)
    id, user_id, movie_id, rating, created_at, updated_at
FROM rating
ORDER BY user_id, movie_id, created_at DESC, id DESC;
