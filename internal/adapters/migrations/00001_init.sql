-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS movie_id_seq;

-- Table Definition
CREATE TABLE "public"."movie" (
    "id" int4 NOT NULL DEFAULT nextval('movie_id_seq'::regclass),
    "created_at" timestamp NOT NULL DEFAULT now(),
    "updated_at" timestamp NOT NULL DEFAULT now(),
    "title" text NOT NULL,
    "runtime" int2 NOT NULL DEFAULT 0,
    "release_date" date,
    "imdb_id" text NOT NULL,
    "imdb_rating" text,
    "overview" text,
    "tagline" text,
    "poster" text,
    "wilhelm" bool,
    "original_title" text,
    "tmdb_id" int4,
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS untitled_table_224_id_seq;

-- Table Definition
CREATE TABLE "public"."award" (
    "id" int4 NOT NULL DEFAULT nextval('untitled_table_224_id_seq'::regclass),
    "imdb_id" text NOT NULL,
    "winner" bool NOT NULL DEFAULT FALSE,
    "name" text NOT NULL,
    "year" text NOT NULL,
    "person" text,
    "person_id" int4,
    "detail" text,
    CONSTRAINT "award_person_id_fkey" FOREIGN KEY ("person_id") REFERENCES "public"."person" ("id"),
    CONSTRAINT "award_imdb_id_fkey" FOREIGN KEY ("imdb_id") REFERENCES "public"."movie" ("imdb_id"),
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS genre_id_seq;

-- Table Definition
CREATE TABLE "public"."genre" (
    "id" int4 NOT NULL DEFAULT nextval('genre_id_seq'::regclass),
    "name" text NOT NULL,
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS language_id_seq;

-- Table Definition
CREATE TABLE "public"."language" (
    "id" int4 NOT NULL DEFAULT nextval('language_id_seq'::regclass),
    "name" text,
    "english_name" text,
    "iso_639_1" text,
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS movie_company_id_seq;

-- Table Definition
CREATE TABLE "public"."movie_company" (
    "id" int4 NOT NULL DEFAULT nextval('movie_company_id_seq'::regclass),
    "movie_id" int4 NOT NULL,
    "company_id" int4 NOT NULL,
    CONSTRAINT "movie_company_company_id_fkey" FOREIGN KEY ("company_id") REFERENCES "public"."production_company" ("id"),
    CONSTRAINT "movie_company_movie_id_fkey" FOREIGN KEY ("movie_id") REFERENCES "public"."movie" ("id"),
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS movie_country_id_seq;

-- Table Definition
CREATE TABLE "public"."movie_country" (
    "id" int4 NOT NULL DEFAULT nextval('movie_country_id_seq'::regclass),
    "movie_id" int4 NOT NULL,
    "country_id" text NOT NULL,
    CONSTRAINT "movie_country_movie_id_fkey" FOREIGN KEY ("movie_id") REFERENCES "public"."movie" ("id"),
    CONSTRAINT "movie_country_country_id_fkey" FOREIGN KEY ("country_id") REFERENCES "public"."production_country" ("id"),
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS movie_genre_id_seq;

-- Table Definition
CREATE TABLE "public"."movie_genre" (
    "id" int4 NOT NULL DEFAULT nextval('movie_genre_id_seq'::regclass),
    "movie_id" int4 NOT NULL,
    "genre_id" int4 NOT NULL,
    CONSTRAINT "movie_genre_genre_id_fkey" FOREIGN KEY ("genre_id") REFERENCES "public"."genre" ("id") ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT "movie_genre_movie_id_fkey" FOREIGN KEY ("movie_id") REFERENCES "public"."movie" ("id") ON DELETE RESTRICT ON UPDATE RESTRICT,
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS movie_language_id_seq;

-- Table Definition
CREATE TABLE "public"."movie_language" (
    "id" int4 NOT NULL DEFAULT nextval('movie_language_id_seq'::regclass),
    "movie_id" int4 NOT NULL,
    "language_id" int4 NOT NULL,
    CONSTRAINT "movie_language_language_id_fkey" FOREIGN KEY ("language_id") REFERENCES "public"."language" ("id"),
    CONSTRAINT "movie_language_movie_id_fkey" FOREIGN KEY ("movie_id") REFERENCES "public"."movie" ("id"),
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS movie_person_id_seq;

DROP TYPE IF EXISTS "public"."job";

CREATE TYPE "public"."job" AS ENUM (
    'cast',
    'composer',
    'director',
    'producer',
    'writer',
    'cinematographer',
    'editor'
);

-- Table Definition
CREATE TABLE "public"."movie_person" (
    "id" int4 NOT NULL DEFAULT nextval('movie_person_id_seq'::regclass),
    "movie_id" int4 NOT NULL,
    "person_id" int4 NOT NULL,
    "job" "public"."job" NOT NULL,
    "character" text,
    CONSTRAINT "movie_person_movie_id_fkey" FOREIGN KEY ("movie_id") REFERENCES "public"."movie" ("id"),
    CONSTRAINT "movie_person_person_id_fkey" FOREIGN KEY ("person_id") REFERENCES "public"."person" ("id") ON DELETE RESTRICT ON UPDATE RESTRICT,
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS movie_series_id_seq;

-- Table Definition
CREATE TABLE "public"."movie_series" (
    "id" int4 NOT NULL DEFAULT nextval('movie_series_id_seq'::regclass),
    "movie_id" int4 NOT NULL,
    "series_id" int4 NOT NULL,
    "number_in_series" int4 NOT NULL,
    CONSTRAINT "movie_series_series_id_fkey" FOREIGN KEY ("series_id") REFERENCES "public"."series" ("id"),
    CONSTRAINT "movie_series_movie_id_fkey" FOREIGN KEY ("movie_id") REFERENCES "public"."movie" ("id"),
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS person_id_seq;

-- Table Definition
CREATE TABLE "public"."person" (
    "id" int4 NOT NULL DEFAULT nextval('person_id_seq'::regclass),
    "name" text NOT NULL,
    "original_id" int4 NOT NULL,
    "popularity" float8,
    "profile_picture" text,
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS production_company_id_seq;

-- Table Definition
CREATE TABLE "public"."production_company" (
    "id" int4 NOT NULL DEFAULT nextval('production_company_id_seq'::regclass),
    "tmdb_id" int4 NOT NULL,
    "name" text NOT NULL,
    "country" text,
    CONSTRAINT "production_company_country_fkey" FOREIGN KEY ("country") REFERENCES "public"."production_country" ("id"),
    PRIMARY KEY ("id")
);

-- Table Definition
CREATE TABLE "public"."production_country" (
    "id" text NOT NULL,
    "name" text NOT NULL,
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS rating_id_seq;

-- Table Definition
CREATE TABLE "public"."rating" (
    "id" int4 NOT NULL DEFAULT nextval('rating_id_seq'::regclass),
    "movie_id" int4 NOT NULL,
    "rating" int2 NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT now(),
    "updated_at" timestamp NOT NULL DEFAULT now(),
    "user_id" int4 NOT NULL,
    CONSTRAINT "rating_movie_id_fkey" FOREIGN KEY ("movie_id") REFERENCES "public"."movie" ("id") ON DELETE RESTRICT ON UPDATE RESTRICT,
    CONSTRAINT "rating_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."user" ("id"),
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS review_id_seq;

-- Table Definition
CREATE TABLE "public"."review" (
    "id" int4 NOT NULL DEFAULT nextval('review_id_seq'::regclass),
    "user_id" int8 NOT NULL,
    "content" text NOT NULL,
    "private" bool NOT NULL DEFAULT TRUE,
    "movie_id" int8 NOT NULL,
    CONSTRAINT "review_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."user" ("id"),
    CONSTRAINT "review_movie_id_fkey" FOREIGN KEY ("movie_id") REFERENCES "public"."movie" ("id"),
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS seen_id_seq;

-- Table Definition
CREATE TABLE "public"."seen" (
    "id" int4 NOT NULL DEFAULT nextval('seen_id_seq'::regclass),
    "movie_id" int4 NOT NULL,
    "date" timestamp NOT NULL DEFAULT now(),
    "user_id" int4 NOT NULL,
    CONSTRAINT "seen_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."user" ("id"),
    CONSTRAINT "seen_movie_id_fkey" FOREIGN KEY ("movie_id") REFERENCES "public"."movie" ("id") ON DELETE RESTRICT ON UPDATE RESTRICT,
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS series_id_seq;

-- Table Definition
CREATE TABLE "public"."series" (
    "id" int4 NOT NULL DEFAULT nextval('series_id_seq'::regclass),
    "name" text NOT NULL,
    PRIMARY KEY ("id")
);

-- Table Definition
CREATE TABLE "public"."series_parents" (
    "series_id" int4 NOT NULL,
    "parent_id" int4 NOT NULL,
    CONSTRAINT "series_parents_parent_id_fkey" FOREIGN KEY ("parent_id") REFERENCES "public"."series" ("id"),
    CONSTRAINT "series_parents_series_id_fkey" FOREIGN KEY ("series_id") REFERENCES "public"."series" ("id"),
    PRIMARY KEY ("series_id", "parent_id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS user_sampleid_seq;

-- Table Definition
CREATE TABLE "public"."user" (
    "id" int4 NOT NULL DEFAULT nextval('user_sampleid_seq'::regclass),
    "created_at" timestamp NOT NULL DEFAULT now(),
    "updated_at" timestamp NOT NULL DEFAULT now(),
    "username" text NOT NULL,
    "password_hash" text NOT NULL,
    "role" text NOT NULL DEFAULT 'user'::text CHECK (ROLE = ANY (ARRAY['user'::text, 'admin'::text])),
    "watch_providers" text,
    PRIMARY KEY ("id")
);

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS watchlist_id_seq;

-- Table Definition
CREATE TABLE "public"."watchlist" (
    "id" int4 NOT NULL DEFAULT nextval('watchlist_id_seq'::regclass),
    "movie_id" int4 NOT NULL,
    "user_id" int4 NOT NULL,
    "created_at" timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY ("id")
);

