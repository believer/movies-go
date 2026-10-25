--
-- PostgreSQL database dump
--


-- Dumped from database version 15.10 (Debian 15.10-1.pgdg120+1)
-- Dumped by pg_dump version 15.10 (Debian 15.10-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: award_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.award_type AS ENUM (
    'academy-award',
    'bafta'
);


--
-- Name: job; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.job AS ENUM (
    'cast',
    'composer',
    'director',
    'producer',
    'writer',
    'cinematographer',
    'editor'
);


--
-- Name: get_person_role_with_seen_json(integer, public.job, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.get_person_role_with_seen_json(pid integer, role public.job, uid integer) RETURNS json
    LANGUAGE sql STABLE
    AS $$
	SELECT
		ARRAY_TO_JSON(ARRAY (
				SELECT
					json_build_object('id', m.id, 'title', m.title, 'releaseDate', to_char(m.release_date, 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), 'seen', s.id IS NOT NULL, 'character', mp."character")
					FROM public.movie_person AS mp
					INNER JOIN public.movie AS m ON m.id = mp.movie_id
					LEFT JOIN ( SELECT DISTINCT ON (movie_id)
							movie_id, id FROM public.seen
						WHERE
							user_id = uid
						ORDER BY
							movie_id, id) AS s ON m.id = s.movie_id
					WHERE
						mp.person_id = pid
						AND mp.job = ROLE
					ORDER BY
						m.release_date DESC))
$$;


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: award; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.award (
    id integer NOT NULL,
    imdb_id text,
    winner boolean DEFAULT false NOT NULL,
    name text NOT NULL,
    year text NOT NULL,
    person text,
    person_id integer,
    detail text,
    type public.award_type DEFAULT 'academy-award'::public.award_type NOT NULL
);


--
-- Name: genre; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.genre (
    id integer NOT NULL,
    name text NOT NULL
);


--
-- Name: genre_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.genre_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: genre_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.genre_id_seq OWNED BY public.genre.id;


--
-- Name: language; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.language (
    id integer NOT NULL,
    name text,
    english_name text,
    iso_639_1 text
);


--
-- Name: language_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.language_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: language_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.language_id_seq OWNED BY public.language.id;


--
-- Name: movie; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.movie (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    title text NOT NULL,
    runtime smallint DEFAULT 0 NOT NULL,
    release_date date,
    imdb_id text NOT NULL,
    imdb_rating text,
    overview text,
    tagline text,
    poster text,
    wilhelm boolean,
    original_title text,
    tmdb_id integer
);


--
-- Name: movie_company; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.movie_company (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    company_id integer NOT NULL
);


--
-- Name: movie_company_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.movie_company_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: movie_company_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.movie_company_id_seq OWNED BY public.movie_company.id;


--
-- Name: movie_country; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.movie_country (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    country_id text NOT NULL
);


--
-- Name: movie_country_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.movie_country_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: movie_country_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.movie_country_id_seq OWNED BY public.movie_country.id;


--
-- Name: movie_genre; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.movie_genre (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    genre_id integer NOT NULL
);


--
-- Name: movie_genre_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.movie_genre_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: movie_genre_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.movie_genre_id_seq OWNED BY public.movie_genre.id;


--
-- Name: movie_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.movie_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: movie_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.movie_id_seq OWNED BY public.movie.id;


--
-- Name: movie_language; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.movie_language (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    language_id integer NOT NULL
);


--
-- Name: movie_language_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.movie_language_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: movie_language_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.movie_language_id_seq OWNED BY public.movie_language.id;


--
-- Name: movie_person; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.movie_person (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    person_id integer NOT NULL,
    job public.job NOT NULL,
    "character" text
);


--
-- Name: movie_person_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.movie_person_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: movie_person_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.movie_person_id_seq OWNED BY public.movie_person.id;


--
-- Name: movie_series; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.movie_series (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    series_id integer NOT NULL,
    number_in_series real NOT NULL
);


--
-- Name: movie_series_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.movie_series_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: movie_series_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.movie_series_id_seq OWNED BY public.movie_series.id;


--
-- Name: now_playing; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.now_playing (
    movie_id integer NOT NULL,
    user_id integer NOT NULL,
    "position" real NOT NULL
);


--
-- Name: official_list; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.official_list (
    id integer NOT NULL,
    slug text NOT NULL,
    name text NOT NULL,
    description text,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    source text NOT NULL
);


--
-- Name: official_list_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.official_list_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: official_list_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.official_list_id_seq OWNED BY public.official_list.id;


--
-- Name: official_list_movie; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.official_list_movie (
    id integer NOT NULL,
    list_id integer NOT NULL,
    movie_id integer NOT NULL,
    rank integer
);


--
-- Name: official_list_movie_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.official_list_movie_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: official_list_movie_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.official_list_movie_id_seq OWNED BY public.official_list_movie.id;


--
-- Name: person; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.person (
    id integer NOT NULL,
    name text NOT NULL,
    original_id integer NOT NULL,
    popularity double precision,
    profile_picture text
);


--
-- Name: person_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.person_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: person_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.person_id_seq OWNED BY public.person.id;


--
-- Name: production_company; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.production_company (
    id integer NOT NULL,
    tmdb_id integer NOT NULL,
    name text NOT NULL,
    country text
);


--
-- Name: production_company_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.production_company_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: production_company_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.production_company_id_seq OWNED BY public.production_company.id;


--
-- Name: production_country; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.production_country (
    id text NOT NULL,
    name text NOT NULL
);


--
-- Name: rating; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.rating (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    rating smallint NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    user_id integer NOT NULL
);


--
-- Name: rating_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.rating_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: rating_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.rating_id_seq OWNED BY public.rating.id;


--
-- Name: review; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.review (
    id integer NOT NULL,
    user_id bigint NOT NULL,
    content text NOT NULL,
    private boolean DEFAULT true NOT NULL,
    movie_id bigint NOT NULL
);


--
-- Name: review_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.review_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: review_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.review_id_seq OWNED BY public.review.id;


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    inserted_at timestamp(0) without time zone
);


--
-- Name: seen; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.seen (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    date timestamp without time zone DEFAULT now() NOT NULL,
    user_id integer NOT NULL
);


--
-- Name: seen_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.seen_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: seen_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.seen_id_seq OWNED BY public.seen.id;


--
-- Name: seen_with; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.seen_with (
    seen_id integer NOT NULL,
    other_user_id integer NOT NULL
);


--
-- Name: series; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.series (
    id integer NOT NULL,
    name text NOT NULL
);


--
-- Name: series_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.series_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: series_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.series_id_seq OWNED BY public.series.id;


--
-- Name: series_parents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.series_parents (
    series_id integer NOT NULL,
    parent_id integer NOT NULL
);


--
-- Name: untitled_table_224_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.untitled_table_224_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: untitled_table_224_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.untitled_table_224_id_seq OWNED BY public.award.id;


--
-- Name: user; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public."user" (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    username text NOT NULL,
    password_hash text NOT NULL,
    role text DEFAULT 'user'::text NOT NULL,
    watch_providers text DEFAULT ''::text NOT NULL,
    name text NOT NULL,
    CONSTRAINT user_role_check CHECK ((role = ANY (ARRAY['user'::text, 'admin'::text])))
);


--
-- Name: user_sampleid_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_sampleid_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_sampleid_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_sampleid_seq OWNED BY public."user".id;


--
-- Name: watchlist; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.watchlist (
    id integer NOT NULL,
    movie_id integer NOT NULL,
    user_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);


--
-- Name: watchlist_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.watchlist_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: watchlist_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.watchlist_id_seq OWNED BY public.watchlist.id;


--
-- Name: award id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.award ALTER COLUMN id SET DEFAULT nextval('public.untitled_table_224_id_seq'::regclass);


--
-- Name: genre id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genre ALTER COLUMN id SET DEFAULT nextval('public.genre_id_seq'::regclass);


--
-- Name: language id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.language ALTER COLUMN id SET DEFAULT nextval('public.language_id_seq'::regclass);


--
-- Name: movie id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie ALTER COLUMN id SET DEFAULT nextval('public.movie_id_seq'::regclass);


--
-- Name: movie_company id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_company ALTER COLUMN id SET DEFAULT nextval('public.movie_company_id_seq'::regclass);


--
-- Name: movie_country id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_country ALTER COLUMN id SET DEFAULT nextval('public.movie_country_id_seq'::regclass);


--
-- Name: movie_genre id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_genre ALTER COLUMN id SET DEFAULT nextval('public.movie_genre_id_seq'::regclass);


--
-- Name: movie_language id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_language ALTER COLUMN id SET DEFAULT nextval('public.movie_language_id_seq'::regclass);


--
-- Name: movie_person id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_person ALTER COLUMN id SET DEFAULT nextval('public.movie_person_id_seq'::regclass);


--
-- Name: movie_series id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_series ALTER COLUMN id SET DEFAULT nextval('public.movie_series_id_seq'::regclass);


--
-- Name: official_list id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.official_list ALTER COLUMN id SET DEFAULT nextval('public.official_list_id_seq'::regclass);


--
-- Name: official_list_movie id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.official_list_movie ALTER COLUMN id SET DEFAULT nextval('public.official_list_movie_id_seq'::regclass);


--
-- Name: person id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person ALTER COLUMN id SET DEFAULT nextval('public.person_id_seq'::regclass);


--
-- Name: production_company id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.production_company ALTER COLUMN id SET DEFAULT nextval('public.production_company_id_seq'::regclass);


--
-- Name: rating id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rating ALTER COLUMN id SET DEFAULT nextval('public.rating_id_seq'::regclass);


--
-- Name: review id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review ALTER COLUMN id SET DEFAULT nextval('public.review_id_seq'::regclass);


--
-- Name: seen id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seen ALTER COLUMN id SET DEFAULT nextval('public.seen_id_seq'::regclass);


--
-- Name: series id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series ALTER COLUMN id SET DEFAULT nextval('public.series_id_seq'::regclass);


--
-- Name: user id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."user" ALTER COLUMN id SET DEFAULT nextval('public.user_sampleid_seq'::regclass);


--
-- Name: watchlist id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.watchlist ALTER COLUMN id SET DEFAULT nextval('public.watchlist_id_seq'::regclass);


--
-- Name: award award_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.award
    ADD CONSTRAINT award_pkey PRIMARY KEY (id);


--
-- Name: genre genre_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genre
    ADD CONSTRAINT genre_name_key UNIQUE (name);


--
-- Name: genre genre_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.genre
    ADD CONSTRAINT genre_pkey PRIMARY KEY (id);


--
-- Name: language language_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.language
    ADD CONSTRAINT language_pkey PRIMARY KEY (id);


--
-- Name: movie_company movie_company_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_company
    ADD CONSTRAINT movie_company_pkey PRIMARY KEY (id);


--
-- Name: movie_country movie_country_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_country
    ADD CONSTRAINT movie_country_pkey PRIMARY KEY (id);


--
-- Name: movie_genre movie_genre_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_genre
    ADD CONSTRAINT movie_genre_pkey PRIMARY KEY (id);


--
-- Name: movie movie_imdb_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie
    ADD CONSTRAINT movie_imdb_id_key UNIQUE (imdb_id);


--
-- Name: movie_language movie_language_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_language
    ADD CONSTRAINT movie_language_pkey PRIMARY KEY (id);


--
-- Name: movie_person movie_person_movie_id_person_id_job_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_person
    ADD CONSTRAINT movie_person_movie_id_person_id_job_key UNIQUE (movie_id, person_id, job);


--
-- Name: movie_person movie_person_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_person
    ADD CONSTRAINT movie_person_pkey PRIMARY KEY (id);


--
-- Name: movie movie_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie
    ADD CONSTRAINT movie_pkey PRIMARY KEY (id);


--
-- Name: movie_series movie_series_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_series
    ADD CONSTRAINT movie_series_pkey PRIMARY KEY (id);


--
-- Name: official_list_movie official_list_movie_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.official_list_movie
    ADD CONSTRAINT official_list_movie_pkey PRIMARY KEY (id);


--
-- Name: official_list official_list_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.official_list
    ADD CONSTRAINT official_list_pkey PRIMARY KEY (id);


--
-- Name: official_list official_list_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.official_list
    ADD CONSTRAINT official_list_slug_key UNIQUE (slug);


--
-- Name: person person_original_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_original_id_key UNIQUE (original_id);


--
-- Name: person person_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.person
    ADD CONSTRAINT person_pkey PRIMARY KEY (id);


--
-- Name: production_company production_company_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.production_company
    ADD CONSTRAINT production_company_pkey PRIMARY KEY (id);


--
-- Name: production_country production_country_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.production_country
    ADD CONSTRAINT production_country_pkey PRIMARY KEY (id);


--
-- Name: rating rating_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT rating_pkey PRIMARY KEY (id);


--
-- Name: review review_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review
    ADD CONSTRAINT review_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: seen seen_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seen
    ADD CONSTRAINT seen_pkey PRIMARY KEY (id);


--
-- Name: series_parents series_parents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series_parents
    ADD CONSTRAINT series_parents_pkey PRIMARY KEY (series_id, parent_id);


--
-- Name: series series_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series
    ADD CONSTRAINT series_pkey PRIMARY KEY (id);


--
-- Name: official_list_movie unique_official_list_movie; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.official_list_movie
    ADD CONSTRAINT unique_official_list_movie UNIQUE (list_id, movie_id);


--
-- Name: user user_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public."user"
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);


--
-- Name: watchlist watchlist_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.watchlist
    ADD CONSTRAINT watchlist_pkey PRIMARY KEY (id);


--
-- Name: watchlist watchlist_user_id_movie_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.watchlist
    ADD CONSTRAINT watchlist_user_id_movie_id_key UNIQUE (user_id, movie_id);


--
-- Name: award_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX award_unique ON public.award USING btree (imdb_id, name, year, person, detail) NULLS NOT DISTINCT;


--
-- Name: idx_movie_person_job; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_movie_person_job ON public.movie_person USING btree (person_id, movie_id, job);


--
-- Name: idx_movie_person_job_person; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_movie_person_job_person ON public.movie_person USING btree (job, person_id);


--
-- Name: idx_movie_person_movie_job; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_movie_person_movie_job ON public.movie_person USING btree (movie_id, job);


--
-- Name: idx_movie_person_movie_person_job; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_movie_person_movie_person_job ON public.movie_person USING btree (movie_id, person_id, job);


--
-- Name: idx_official_list_movie_order; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_official_list_movie_order ON public.official_list_movie USING btree (list_id, rank);


--
-- Name: idx_person_group; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_person_group ON public.person USING btree (id, name);


--
-- Name: idx_person_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_person_id ON public.movie_person USING btree (person_id);


--
-- Name: idx_rating_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_rating_user_id ON public.rating USING btree (user_id);


--
-- Name: idx_seen_user_filtered; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_seen_user_filtered ON public.seen USING btree (movie_id) WHERE (user_id = 1);


--
-- Name: idx_seen_user_year; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_seen_user_year ON public.seen USING btree (user_id, EXTRACT(year FROM date));


--
-- Name: language_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX language_unique ON public.language USING btree (name);


--
-- Name: movie_company_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX movie_company_unique ON public.movie_company USING btree (movie_id, company_id);


--
-- Name: movie_country_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX movie_country_unique ON public.movie_country USING btree (movie_id, country_id);


--
-- Name: movie_created_at_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movie_created_at_index ON public.movie USING btree (created_at DESC NULLS LAST);


--
-- Name: movie_genre_genre_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movie_genre_genre_id_idx ON public.movie_genre USING btree (genre_id);


--
-- Name: movie_genre_movie_id_genre_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX movie_genre_movie_id_genre_id_key ON public.movie_genre USING btree (genre_id, movie_id);


--
-- Name: movie_genre_movie_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX movie_genre_movie_id_idx ON public.movie_genre USING btree (movie_id);


--
-- Name: movie_id_user_pkey; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX movie_id_user_pkey ON public.now_playing USING btree (movie_id, user_id);


--
-- Name: movie_language_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX movie_language_unique ON public.movie_language USING btree (movie_id, language_id);


--
-- Name: movie_series_movie_series; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX movie_series_movie_series ON public.movie_series USING btree (movie_id, series_id, number_in_series);


--
-- Name: production_tmdb_id_name_pkey; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX production_tmdb_id_name_pkey ON public.production_company USING btree (tmdb_id, name);


--
-- Name: rating_movie_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX rating_movie_id_idx ON public.rating USING btree (movie_id);


--
-- Name: review_user_movie; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX review_user_movie ON public.review USING btree (movie_id, user_id);


--
-- Name: seen_created_at_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX seen_created_at_index ON public.seen USING btree (date DESC NULLS LAST);


--
-- Name: seen_movie_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX seen_movie_id_idx ON public.seen USING btree (movie_id);


--
-- Name: seen_movie_id_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX seen_movie_id_user_id_idx ON public.seen USING btree (user_id, movie_id);


--
-- Name: seen_other_user; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX seen_other_user ON public.seen_with USING btree (seen_id, other_user_id);


--
-- Name: seen_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX seen_user_id_idx ON public.seen USING btree (user_id);

ALTER TABLE public.seen CLUSTER ON seen_user_id_idx;


--
-- Name: series_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX series_name ON public.series USING btree (name);


--
-- Name: user_username_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX user_username_key ON public."user" USING btree (username);


--
-- Name: movie update_ab_time; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER update_ab_time BEFORE UPDATE ON public.movie FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: award award_imdb_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.award
    ADD CONSTRAINT award_imdb_id_fkey FOREIGN KEY (imdb_id) REFERENCES public.movie(imdb_id);


--
-- Name: award award_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.award
    ADD CONSTRAINT award_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(id);


--
-- Name: movie_company movie_company_company_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_company
    ADD CONSTRAINT movie_company_company_id_fkey FOREIGN KEY (company_id) REFERENCES public.production_company(id);


--
-- Name: movie_company movie_company_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_company
    ADD CONSTRAINT movie_company_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id);


--
-- Name: movie_country movie_country_country_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_country
    ADD CONSTRAINT movie_country_country_id_fkey FOREIGN KEY (country_id) REFERENCES public.production_country(id);


--
-- Name: movie_country movie_country_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_country
    ADD CONSTRAINT movie_country_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id);


--
-- Name: movie_genre movie_genre_genre_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_genre
    ADD CONSTRAINT movie_genre_genre_id_fkey FOREIGN KEY (genre_id) REFERENCES public.genre(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: movie_genre movie_genre_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_genre
    ADD CONSTRAINT movie_genre_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: movie_language movie_language_language_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_language
    ADD CONSTRAINT movie_language_language_id_fkey FOREIGN KEY (language_id) REFERENCES public.language(id);


--
-- Name: movie_language movie_language_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_language
    ADD CONSTRAINT movie_language_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id);


--
-- Name: movie_person movie_person_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_person
    ADD CONSTRAINT movie_person_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id);


--
-- Name: movie_person movie_person_person_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_person
    ADD CONSTRAINT movie_person_person_id_fkey FOREIGN KEY (person_id) REFERENCES public.person(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: movie_series movie_series_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_series
    ADD CONSTRAINT movie_series_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id);


--
-- Name: movie_series movie_series_series_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.movie_series
    ADD CONSTRAINT movie_series_series_id_fkey FOREIGN KEY (series_id) REFERENCES public.series(id);


--
-- Name: official_list_movie official_list_movie_list_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.official_list_movie
    ADD CONSTRAINT official_list_movie_list_id_fkey FOREIGN KEY (list_id) REFERENCES public.official_list(id) ON DELETE CASCADE;


--
-- Name: official_list_movie official_list_movie_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.official_list_movie
    ADD CONSTRAINT official_list_movie_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id) ON DELETE CASCADE;


--
-- Name: production_company production_company_country_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.production_company
    ADD CONSTRAINT production_company_country_fkey FOREIGN KEY (country) REFERENCES public.production_country(id);


--
-- Name: rating rating_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT rating_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: rating rating_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rating
    ADD CONSTRAINT rating_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id);


--
-- Name: review review_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review
    ADD CONSTRAINT review_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id);


--
-- Name: review review_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.review
    ADD CONSTRAINT review_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id);


--
-- Name: seen seen_movie_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seen
    ADD CONSTRAINT seen_movie_id_fkey FOREIGN KEY (movie_id) REFERENCES public.movie(id) ON UPDATE RESTRICT ON DELETE RESTRICT;


--
-- Name: seen seen_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seen
    ADD CONSTRAINT seen_user_id_fkey FOREIGN KEY (user_id) REFERENCES public."user"(id);


--
-- Name: seen_with seen_with_other_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seen_with
    ADD CONSTRAINT seen_with_other_user_id_fkey FOREIGN KEY (other_user_id) REFERENCES public."user"(id);


--
-- Name: seen_with seen_with_seen_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.seen_with
    ADD CONSTRAINT seen_with_seen_id_fkey FOREIGN KEY (seen_id) REFERENCES public.seen(id);


--
-- Name: series_parents series_parents_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series_parents
    ADD CONSTRAINT series_parents_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.series(id);


--
-- Name: series_parents series_parents_series_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.series_parents
    ADD CONSTRAINT series_parents_series_id_fkey FOREIGN KEY (series_id) REFERENCES public.series(id);


--
-- PostgreSQL database dump complete
--


