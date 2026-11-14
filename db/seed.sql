-- Fixed data for the E2E database, applied on top of db/schema.sql by
-- make e2e-db. E2E tests assert only on rows from this file, so a test that
-- needs more data adds it here. Movies, people and awards are real; Watches,
-- Ratings and Reviews are made up and sit in fixed past years.
--
-- Rows use explicit ids so tests and readers can refer to them. The sequences
-- are moved past them at the end, so rows the app inserts don't collide.

-- The schema dump empties search_path when it runs in the same session
SET search_path = public;

-- Users
-- =====================================================
-- User 1 is the user the app falls back to without a token. Both passwords
-- are "password".

INSERT INTO "user" (id, username, password_hash, name, role) VALUES
    (1, 'admin', '$2a$10$X..zDsQcVpaTDGNWUapChOjvBfbDONb1N.vkJucZWkAFmY8T32dyq', 'Admin', 'admin'),
    (2, 'friend', '$2a$10$X..zDsQcVpaTDGNWUapChOjvBfbDONb1N.vkJucZWkAFmY8T32dyq', 'Friend', 'user');

-- Metadata
-- =====================================================

INSERT INTO genre (id, name) VALUES
    (1, 'Action'),
    (2, 'Adventure'),
    (3, 'Animation'),
    (4, 'Comedy'),
    (5, 'Crime'),
    (6, 'Drama'),
    (7, 'Fantasy'),
    (8, 'Romance'),
    (9, 'Science Fiction'),
    (10, 'Thriller');

INSERT INTO language (id, name, english_name, iso_639_1) VALUES
    (1, 'English', 'English', 'en'),
    (2, 'Français', 'French', 'fr'),
    (3, '日本語', 'Japanese', 'ja'),
    (4, '한국어/조선말', 'Korean', 'ko');

INSERT INTO production_country (id, name) VALUES
    ('US', 'United States of America'),
    ('GB', 'United Kingdom'),
    ('FR', 'France'),
    ('JP', 'Japan'),
    ('KR', 'South Korea');

INSERT INTO production_company (id, tmdb_id, name, country) VALUES
    (1, 174, 'Warner Bros. Pictures', 'US'),
    (2, 923, 'Legendary Pictures', 'US'),
    (3, 9996, 'Syncopy', 'GB'),
    (4, 56, 'Amblin Entertainment', 'US'),
    (5, 420, 'Marvel Studios', 'US'),
    (6, 10342, 'Studio Ghibli', 'JP'),
    (7, 2, 'Walt Disney Pictures', 'US');

-- Movies
-- =====================================================

INSERT INTO movie (id, title, original_title, release_date, runtime, imdb_id, tmdb_id, imdb_rating, tagline, overview, created_at, updated_at) VALUES
    (1, 'Batman Begins', 'Batman Begins', '2005-06-10', 140, 'tt0372784', 272, '8.2', 'Evil fears the knight.', 'Driven by tragedy, billionaire Bruce Wayne dedicates his life to uncovering and defeating the corruption that plagues his home, Gotham City.', '2019-01-01 12:00', '2019-01-01 12:00'),
    (2, 'The Dark Knight', 'The Dark Knight', '2008-07-16', 152, 'tt0468569', 155, '9.0', 'Welcome to a world without rules.', 'Batman raises the stakes in his war on crime, until a criminal mastermind known as the Joker throws Gotham into anarchy.', '2019-01-01 12:00', '2019-01-01 12:00'),
    (3, 'Inception', 'Inception', '2010-07-15', 148, 'tt1375666', 27205, '8.8', 'Your mind is the scene of the crime.', 'Cobb, a skilled thief who steals secrets from deep within the subconscious, is offered a chance to have his past crimes forgiven.', '2019-01-01 12:00', '2019-01-01 12:00'),
    (4, 'Interstellar', 'Interstellar', '2014-11-05', 169, 'tt0816692', 157336, '8.7', 'Mankind was born on Earth. It was never meant to die here.', 'A team of explorers travel through a wormhole in space in an attempt to ensure humanity''s survival.', '2019-01-01 12:00', '2019-01-01 12:00'),
    (5, 'Dune', 'Dune', '2021-09-15', 155, 'tt1160419', 438631, '8.0', 'It begins.', 'Paul Atreides, a brilliant and gifted young man born into a great destiny beyond his understanding, must travel to the most dangerous planet in the universe.', '2021-10-01 12:00', '2021-10-01 12:00'),
    (6, 'Jurassic Park', 'Jurassic Park', '1993-06-11', 127, 'tt0107290', 329, '8.2', 'An adventure 65 million years in the making.', 'A wealthy entrepreneur secretly creates a theme park featuring living dinosaurs drawn from prehistoric DNA.', '2019-01-01 12:00', '2019-01-01 12:00'),
    (7, 'Ant-Man', 'Ant-Man', '2015-07-14', 117, 'tt0478970', 102899, '7.3', 'Heroes don''t get any bigger.', 'Armed with the astonishing ability to shrink in scale but increase in strength, master thief Scott Lang must embrace his inner-hero.', '2020-01-01 12:00', '2020-01-01 12:00'),
    (8, 'Kingsman: The Secret Service', 'Kingsman: The Secret Service', '2015-01-24', 129, 'tt2802144', 207703, '7.7', 'Manners maketh man.', 'The story of a super-secret spy organization that recruits an unrefined but promising street kid into the agency''s ultra-competitive training program.', '2020-01-01 12:00', '2020-01-01 12:00'),
    (9, 'Amélie', 'Le Fabuleux Destin d''Amélie Poulain', '2001-04-25', 122, 'tt0211915', 194, '8.3', 'One person can change your life forever.', 'At a tiny Parisian café, the adorable yet painfully shy Amélie accidentally discovers a gift for helping others.', '2020-01-01 12:00', '2020-01-01 12:00'),
    (10, 'Spirited Away', '千と千尋の神隠し', '2001-07-20', 125, 'tt0245429', 129, '8.6', NULL, 'A young girl, Chihiro, becomes trapped in a strange new world of spirits.', '2020-01-01 12:00', '2020-01-01 12:00'),
    (11, 'Parasite', '기생충', '2019-05-30', 133, 'tt6751668', 496243, '8.5', 'Act like you own the place.', 'All unemployed, Ki-taek''s family takes peculiar interest in the wealthy and glamorous Parks for their livelihood.', '2020-01-01 12:00', '2020-01-01 12:00'),
    (12, 'The Lion King', 'The Lion King', '1994-06-24', 89, 'tt0110357', 8587, '8.5', 'The greatest adventure of all is finding our place in the circle of life.', 'Young lion prince Simba, eager to one day become king of the Pride Lands, grows up under the watchful eye of his father Mufasa.', '2021-01-01 12:00', '2021-01-01 12:00');

INSERT INTO movie_genre (movie_id, genre_id) VALUES
    (1, 1), (1, 5), (1, 6),
    (2, 1), (2, 5), (2, 6), (2, 10),
    (3, 1), (3, 9), (3, 2),
    (4, 2), (4, 6), (4, 9),
    (5, 9), (5, 2),
    (6, 2), (6, 9),
    (7, 1), (7, 9), (7, 4),
    (8, 5), (8, 4), (8, 1), (8, 2),
    (9, 4), (9, 8),
    (10, 3), (10, 7),
    (11, 4), (11, 10), (11, 6),
    (12, 3), (12, 6), (12, 2);

INSERT INTO movie_language (movie_id, language_id) VALUES
    (1, 1), (2, 1), (3, 1), (3, 3), (3, 2), (4, 1), (5, 1),
    (6, 1), (7, 1), (8, 1), (9, 2), (10, 3), (11, 4), (12, 1);

INSERT INTO movie_country (movie_id, country_id) VALUES
    (1, 'US'), (1, 'GB'), (2, 'US'), (2, 'GB'), (3, 'US'), (3, 'GB'),
    (4, 'US'), (4, 'GB'), (5, 'US'), (6, 'US'), (7, 'US'), (8, 'GB'),
    (8, 'US'), (9, 'FR'), (10, 'JP'), (11, 'KR'), (12, 'US');

INSERT INTO movie_company (movie_id, company_id) VALUES
    (1, 1), (1, 2), (1, 3), (2, 1), (2, 2), (2, 3), (3, 1), (3, 2), (3, 3),
    (4, 1), (4, 2), (4, 3), (5, 1), (5, 2), (6, 4), (7, 5), (10, 6), (12, 7);

INSERT INTO series (id, name) VALUES
    (1, 'The Dark Knight Trilogy');

INSERT INTO movie_series (movie_id, series_id, number_in_series) VALUES
    (1, 1, 1),
    (2, 1, 2);

-- People
-- =====================================================
-- One known match per search prefix: cast:paul, director:steven,
-- producer:matt and composer:hans each find a single person.

INSERT INTO person (id, name, original_id, popularity) VALUES
    (1, 'Christian Bale', 3894, 20),
    (2, 'Heath Ledger', 1810, 15),
    (3, 'Christopher Nolan', 525, 18),
    (4, 'Emma Thomas', 556, 5),
    (5, 'Hans Zimmer', 947, 10),
    (6, 'Leonardo DiCaprio', 6193, 30),
    (7, 'Matthew McConaughey', 10297, 22),
    (8, 'Timothée Chalamet', 1190668, 40),
    (9, 'Denis Villeneuve', 137427, 12),
    (10, 'Steven Spielberg', 488, 14),
    (11, 'Kathleen Kennedy', 489, 4),
    (12, 'John Williams', 491, 8),
    (13, 'Paul Rudd', 22226, 25),
    (14, 'Taron Egerton', 1197313, 16),
    (15, 'Matthew Vaughn', 957, 6),
    (16, 'Audrey Tautou', 3842, 9),
    (17, 'Hayao Miyazaki', 608, 11),
    (18, 'Bong Joon-ho', 21684, 13);

INSERT INTO movie_person (movie_id, person_id, job, "character") VALUES
    (1, 1, 'cast', 'Bruce Wayne / Batman'),
    (1, 3, 'director', NULL),
    (1, 4, 'producer', NULL),
    (1, 5, 'composer', NULL),
    (2, 1, 'cast', 'Bruce Wayne / Batman'),
    (2, 2, 'cast', 'Joker'),
    (2, 3, 'director', NULL),
    (2, 4, 'producer', NULL),
    (2, 5, 'composer', NULL),
    (3, 6, 'cast', 'Cobb'),
    (3, 3, 'director', NULL),
    (3, 4, 'producer', NULL),
    (3, 5, 'composer', NULL),
    (4, 7, 'cast', 'Cooper'),
    (4, 8, 'cast', 'Tom'),
    (4, 3, 'director', NULL),
    (4, 4, 'producer', NULL),
    (4, 5, 'composer', NULL),
    (5, 8, 'cast', 'Paul Atreides'),
    (5, 9, 'director', NULL),
    (5, 5, 'composer', NULL),
    (6, 10, 'director', NULL),
    (6, 11, 'producer', NULL),
    (6, 12, 'composer', NULL),
    (7, 13, 'cast', 'Scott Lang / Ant-Man'),
    (8, 14, 'cast', 'Eggsy'),
    (8, 15, 'director', NULL),
    (8, 15, 'producer', NULL),
    (9, 16, 'cast', 'Amélie Poulain'),
    (10, 17, 'director', NULL),
    (11, 18, 'director', NULL),
    (12, 5, 'composer', NULL);

-- Awards
-- =====================================================

INSERT INTO award (id, imdb_id, name, year, person, person_id, detail, winner, type) VALUES
    (1, 'tt1160419', 'Music (Original Score)', '2021', 'Hans Zimmer', 5, NULL, true, 'academy-award'),
    (2, 'tt0816692', 'Music (Original Score)', '2014', 'Hans Zimmer', 5, NULL, false, 'academy-award'),
    (3, 'tt1375666', 'Music (Original Score)', '2010', 'Hans Zimmer', 5, NULL, false, 'academy-award'),
    (4, 'tt0468569', 'Actor in a Supporting Role', '2008', 'Heath Ledger', 2, 'Joker', true, 'academy-award'),
    (5, 'tt6751668', 'Best Picture', '2019', NULL, NULL, NULL, true, 'academy-award'),
    (6, 'tt6751668', 'Directing', '2019', 'Bong Joon-ho', 18, NULL, true, 'academy-award'),
    (7, 'tt1160419', 'Best Picture', '2021', NULL, NULL, NULL, false, 'academy-award'),
    (8, 'tt1160419', 'Original Score', '2022', 'Hans Zimmer', 5, NULL, true, 'bafta');

-- Watches, Ratings and Reviews
-- =====================================================
-- User 1 watches in 2019 to 2022, and rewatches The Dark Knight. User 2 has
-- their own Watches, some of movies user 1 hasn't seen, so one user's data
-- leaking into the other's views shows up.

INSERT INTO seen (id, movie_id, user_id, date) VALUES
    (1, 1, 1, '2019-03-02 20:00'),
    (2, 2, 1, '2019-03-09 20:00'),
    (3, 3, 1, '2019-06-15 21:00'),
    (4, 6, 1, '2019-08-10 19:30'),
    (5, 4, 1, '2020-02-14 20:00'),
    (6, 9, 1, '2020-05-01 18:00'),
    (7, 10, 1, '2020-07-04 16:00'),
    (8, 11, 1, '2020-11-21 21:00'),
    (9, 7, 1, '2021-01-16 20:00'),
    (10, 5, 1, '2021-10-23 19:00'),
    (11, 2, 1, '2021-12-30 20:00'),
    (12, 8, 1, '2022-04-09 20:00'),
    (13, 2, 2, '2019-03-09 20:00'),
    (14, 12, 2, '2021-06-12 15:00'),
    (15, 5, 2, '2022-01-08 19:00');

-- User 2 watched The Dark Knight with user 1
INSERT INTO seen_with (seen_id, other_user_id) VALUES
    (2, 2);

-- User 1 rerates The Dark Knight lower, so "latest rating wins" differs from
-- "highest rating wins"
INSERT INTO rating (id, movie_id, user_id, rating, created_at, updated_at) VALUES
    (1, 1, 1, 8, '2019-03-02 22:30', '2019-03-02 22:30'),
    (2, 2, 1, 10, '2019-03-09 22:40', '2019-03-09 22:40'),
    (3, 3, 1, 9, '2019-06-15 23:30', '2019-06-15 23:30'),
    (4, 6, 1, 8, '2019-08-10 21:40', '2019-08-10 21:40'),
    (5, 4, 1, 9, '2020-02-14 22:50', '2020-02-14 22:50'),
    (6, 9, 1, 7, '2020-05-01 20:05', '2020-05-01 20:05'),
    (7, 10, 1, 10, '2020-07-04 18:05', '2020-07-04 18:05'),
    (8, 11, 1, 9, '2020-11-21 23:15', '2020-11-21 23:15'),
    (9, 7, 1, 6, '2021-01-16 22:00', '2021-01-16 22:00'),
    (10, 5, 1, 8, '2021-10-23 21:40', '2021-10-23 21:40'),
    (11, 2, 1, 9, '2021-12-30 22:40', '2021-12-30 22:40'),
    (12, 8, 1, 7, '2022-04-09 22:10', '2022-04-09 22:10'),
    (13, 2, 2, 6, '2019-03-09 22:40', '2019-03-09 22:40'),
    (14, 12, 2, 10, '2021-06-12 16:30', '2021-06-12 16:30'),
    (15, 5, 2, 5, '2022-01-08 21:40', '2022-01-08 21:40');

INSERT INTO review (id, movie_id, user_id, content, private) VALUES
    (1, 2, 1, 'Heath Ledger steals every scene he is in.', false),
    (2, 5, 1, 'The sound design alone is worth the ticket.', true);

-- Watchlist
-- =====================================================

INSERT INTO watchlist (id, movie_id, user_id, created_at) VALUES
    (1, 12, 1, '2022-05-01 12:00'),
    (2, 3, 2, '2022-05-01 12:00');

-- Sequences
-- =====================================================

SELECT setval(pg_get_serial_sequence('public."user"', 'id'), (SELECT max(id) FROM "user"));
SELECT setval(pg_get_serial_sequence('public.genre', 'id'), (SELECT max(id) FROM genre));
SELECT setval(pg_get_serial_sequence('public.language', 'id'), (SELECT max(id) FROM language));
SELECT setval(pg_get_serial_sequence('public.production_company', 'id'), (SELECT max(id) FROM production_company));
SELECT setval(pg_get_serial_sequence('public.movie', 'id'), (SELECT max(id) FROM movie));
SELECT setval(pg_get_serial_sequence('public.series', 'id'), (SELECT max(id) FROM series));
SELECT setval(pg_get_serial_sequence('public.person', 'id'), (SELECT max(id) FROM person));
SELECT setval(pg_get_serial_sequence('public.award', 'id'), (SELECT max(id) FROM award));
SELECT setval(pg_get_serial_sequence('public.seen', 'id'), (SELECT max(id) FROM seen));
SELECT setval(pg_get_serial_sequence('public.rating', 'id'), (SELECT max(id) FROM rating));
SELECT setval(pg_get_serial_sequence('public.review', 'id'), (SELECT max(id) FROM review));
SELECT setval(pg_get_serial_sequence('public.watchlist', 'id'), (SELECT max(id) FROM watchlist));
