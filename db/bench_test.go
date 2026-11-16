package db

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// The benchmarks time each repository method that takes a user ID against a
// copy of the production backup, built with make bench-db and run with make
// bench. docs/agents/performance.md has the procedure.

// benchArgs holds representative arguments from the backup: the user with the
// most Watches, and the largest genre, list, series, person and so on.
type benchArgs struct {
	UserID string `db:"user_id"`
	// The user's most watched movie
	MovieID int `db:"movie_id"`
	// The user's most watched movie with a single Rating, so UpdateRating
	// rewrites no history
	RatedMovieID int `db:"rated_movie_id"`
	// The user's most watched movie without a Review from them
	UnreviewedMovieID int `db:"unreviewed_movie_id"`
	// The user's movie with the longest Review
	ReviewedMovieID int `db:"reviewed_movie_id"`
	// The user's most watched movie they're not playing
	UnplayedMovieID int `db:"unplayed_movie_id"`
	// The user's Watch with the most friends
	SeenID string `db:"seen_id"`
	// The user's oldest watchlist entry
	WatchlistMovieID int    `db:"watchlist_movie_id"`
	GenreID          string `db:"genre_id"`
	LanguageID       string `db:"language_id"`
	CompanyID        string `db:"company_id"`
	CountryID        string `db:"country_id"`
	ListID           string `db:"list_id"`
	// The series with the most movies, counting those of its sub-series
	SeriesID string `db:"series_id"`
	PersonID string `db:"person_id"`
	// The year the user watched the most movies in
	WatchYear string `db:"watch_year"`
	// The year the most movies were released in
	ReleaseYear string `db:"release_year"`
	// The user's most common rating
	Rating int `db:"rating"`
}

// benchArgsQuery picks benchArgs. Ties go to the lowest ID, so two copies of
// the backup pick the same arguments.
const benchArgsQuery = `
WITH RECURSIVE bench_user AS (
    SELECT
        user_id AS id
    FROM
        seen
    GROUP BY
        user_id
    ORDER BY
        count(*) DESC,
        user_id
    LIMIT 1
),
user_movie AS (
    SELECT
        movie_id,
        count(*) AS watches
    FROM
        seen
    WHERE
        user_id = (SELECT id FROM bench_user)
    GROUP BY
        movie_id
),
series_tree AS (
    SELECT
        id AS root_id,
        id AS series_id
    FROM
        series
    UNION ALL
    SELECT
        t.root_id,
        sp.series_id
    FROM
        series_tree t
        INNER JOIN series_parents sp ON sp.parent_id = t.series_id
)
SELECT
    (SELECT id FROM bench_user)::text AS user_id,
    (SELECT movie_id FROM user_movie ORDER BY watches DESC, movie_id LIMIT 1) AS movie_id,
    (SELECT um.movie_id FROM user_movie um
        INNER JOIN rating r ON r.movie_id = um.movie_id AND r.user_id = (SELECT id FROM bench_user)
        GROUP BY um.movie_id, um.watches
        HAVING count(*) = 1
        ORDER BY um.watches DESC, um.movie_id LIMIT 1) AS rated_movie_id,
    (SELECT um.movie_id FROM user_movie um
        WHERE NOT EXISTS (SELECT FROM review WHERE movie_id = um.movie_id AND user_id = (SELECT id FROM bench_user))
        ORDER BY um.watches DESC, um.movie_id LIMIT 1) AS unreviewed_movie_id,
    (SELECT movie_id FROM review WHERE user_id = (SELECT id FROM bench_user)
        ORDER BY length(content) DESC, movie_id LIMIT 1)::int AS reviewed_movie_id,
    (SELECT um.movie_id FROM user_movie um
        WHERE NOT EXISTS (SELECT FROM now_playing WHERE movie_id = um.movie_id AND user_id = (SELECT id FROM bench_user))
        ORDER BY um.watches DESC, um.movie_id LIMIT 1) AS unplayed_movie_id,
    (SELECT s.id FROM seen s
        LEFT JOIN seen_with sw ON sw.seen_id = s.id
        WHERE s.user_id = (SELECT id FROM bench_user)
        GROUP BY s.id
        ORDER BY count(sw.seen_id) DESC, s.id LIMIT 1)::text AS seen_id,
    (SELECT movie_id FROM watchlist WHERE user_id = (SELECT id FROM bench_user)
        ORDER BY created_at, movie_id LIMIT 1) AS watchlist_movie_id,
    (SELECT genre_id FROM movie_genre GROUP BY genre_id ORDER BY count(*) DESC, genre_id LIMIT 1)::text AS genre_id,
    (SELECT language_id FROM movie_language GROUP BY language_id ORDER BY count(*) DESC, language_id LIMIT 1)::text AS language_id,
    (SELECT company_id FROM movie_company GROUP BY company_id ORDER BY count(*) DESC, company_id LIMIT 1)::text AS company_id,
    (SELECT country_id FROM movie_country GROUP BY country_id ORDER BY count(*) DESC, country_id LIMIT 1) AS country_id,
    (SELECT list_id FROM official_list_movie GROUP BY list_id ORDER BY count(*) DESC, list_id LIMIT 1)::text AS list_id,
    (SELECT t.root_id FROM series_tree t
        INNER JOIN movie_series ms ON ms.series_id = t.series_id
        GROUP BY t.root_id
        ORDER BY count(*) DESC, t.root_id LIMIT 1)::text AS series_id,
    (SELECT person_id FROM movie_person GROUP BY person_id ORDER BY count(*) DESC, person_id LIMIT 1)::text AS person_id,
    -- Watches before 2012 are a bulk import, dated to a few days of 2011
    (SELECT date_part('year', date)::text FROM seen
        WHERE user_id = (SELECT id FROM bench_user) AND date >= '2012-01-01'
        GROUP BY 1 ORDER BY count(*) DESC, 1 LIMIT 1) AS watch_year,
    (SELECT date_part('year', release_date)::text FROM movie
        WHERE release_date IS NOT NULL
        GROUP BY 1 ORDER BY count(*) DESC, 1 LIMIT 1) AS release_year,
    (SELECT rating FROM latest_rating WHERE user_id = (SELECT id FROM bench_user)
        GROUP BY rating ORDER BY count(*) DESC, rating LIMIT 1)::int AS rating
`

type benchEnv struct {
	db   *sqlx.DB
	args benchArgs
}

// openBenchEnv connects once per run, since go test calls each benchmark
// several times to find b.N
var openBenchEnv = sync.OnceValues(func() (benchEnv, error) {
	url := os.Getenv("BENCH_DATABASE_URL")
	db, err := sqlx.Connect("postgres", url)

	if err != nil {
		return benchEnv{}, fmt.Errorf("connect to BENCH_DATABASE_URL (run make bench-db first): %w", err)
	}

	var args benchArgs

	if err := db.Get(&args, benchArgsQuery); err != nil {
		return benchEnv{}, fmt.Errorf("find representative arguments, is BENCH_DATABASE_URL a copy of the production backup? %w", err)
	}

	return benchEnv{db, args}, nil
})

// benchDB returns the bench database and its representative arguments, or
// skips the benchmark when BENCH_DATABASE_URL is unset.
func benchDB(b *testing.B) (*sqlx.DB, benchArgs) {
	b.Helper()

	if os.Getenv("BENCH_DATABASE_URL") == "" {
		b.Skip("BENCH_DATABASE_URL not set: run make bench-db, then make bench")
	}

	env, err := openBenchEnv()

	if err != nil {
		b.Fatal(err)
	}

	return env.db, env.args
}

// benchCase times run. before and after, when set, run off the timer around
// each call: a write uses them to put the data back, so every iteration and
// every later run sees the same rows.
type benchCase struct {
	name   string
	before func(b *testing.B)
	run    func() error
	after  func(b *testing.B)
}

func runBench(b *testing.B, cases []benchCase) {
	b.Helper()

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			for b.Loop() {
				if c.before != nil {
					b.StopTimer()
					c.before(b)
					b.StartTimer()
				}

				if err := c.run(); err != nil {
					b.Fatal(err)
				}

				if c.after != nil {
					b.StopTimer()
					c.after(b)
					b.StartTimer()
				}
			}
		})
	}
}

func benchExec(b *testing.B, db *sqlx.DB, query string, args ...any) {
	b.Helper()

	if _, err := db.Exec(query, args...); err != nil {
		b.Fatal(err)
	}
}

// benchTx gives a case a transaction that's rolled back after each call
func benchTx(db *sqlx.DB, tx **sqlx.Tx) (before, after func(b *testing.B)) {
	before = func(b *testing.B) {
		var err error

		if *tx, err = db.Beginx(); err != nil {
			b.Fatal(err)
		}
	}

	after = func(b *testing.B) {
		if err := (*tx).Rollback(); err != nil {
			b.Fatal(err)
		}
	}

	return before, after
}

func BenchmarkAwards(b *testing.B) {
	db, a := benchDB(b)
	r := NewAwardsRepository(db)

	// One award is the largest result: every movie with an award has one
	runBench(b, []benchCase{
		{name: "GetByNominations", run: func() error { _, err := r.GetByNominations(a.UserID, 1, "academy-award"); return err }},
		{name: "GetByWins", run: func() error { _, err := r.GetByWins(a.UserID, 1, "academy-award"); return err }},
	})
}

func BenchmarkFeed(b *testing.B) {
	db, a := benchDB(b)
	r := NewFeedRepository(db)

	runBench(b, []benchCase{
		{name: "SearchMoviesByRating", run: func() error {
			_, err := r.SearchMoviesByRating(strconv.Itoa(a.Rating), a.UserID)
			return err
		}},
		{name: "GetFeedMovies", run: func() error { _, err := r.GetFeedMovies(a.UserID, 0); return err }},
	})
}

func BenchmarkGenre(b *testing.B) {
	db, a := benchDB(b)
	r := NewGenreRepository(db)

	runBench(b, []benchCase{
		{name: "GetGenreMovies", run: func() error { _, err := r.GetGenreMovies(a.GenreID, a.UserID, 0); return err }},
		{name: "GetGenreStats", run: func() error { _, err := r.GetGenreStats(a.UserID, "All"); return err }},
	})
}

func BenchmarkLanguage(b *testing.B) {
	db, a := benchDB(b)
	r := NewLanguageRepository(db)

	runBench(b, []benchCase{
		{name: "GetLanguageMovies", run: func() error { _, err := r.GetLanguageMovies(a.LanguageID, a.UserID, 0); return err }},
		{name: "GetLanguageStats", run: func() error { _, err := r.GetLanguageStats(a.UserID, "All"); return err }},
	})
}

func BenchmarkList(b *testing.B) {
	db, a := benchDB(b)
	r := NewListRepository(db)

	runBench(b, []benchCase{
		{name: "GetLists", run: func() error { _, err := r.GetLists(a.UserID); return err }},
		{name: "GetListMovies", run: func() error { _, err := r.GetListMovies(a.ListID, a.UserID); return err }},
	})
}

func BenchmarkMovie(b *testing.B) {
	db, a := benchDB(b)
	r := NewMovieRepository(db)
	movieID := strconv.Itoa(a.MovieID)

	// UpdateSeenMovie writes the Watch's own date and friends back
	var watch struct {
		Date    time.Time      `db:"date"`
		Friends pq.StringArray `db:"friends"`
	}

	err := db.Get(&watch, `
SELECT
    s.date,
    COALESCE(ARRAY_AGG(sw.other_user_id::text ORDER BY sw.other_user_id) FILTER (WHERE sw.other_user_id IS NOT NULL), '{}') AS friends
FROM
    seen s
    LEFT JOIN seen_with sw ON sw.seen_id = s.id
WHERE
    s.id = $1
    AND s.user_id = $2
GROUP BY
    s.id`, a.SeenID, a.UserID)

	if err != nil {
		b.Fatal(err)
	}

	var rating int

	if err := db.Get(&rating, `SELECT rating FROM latest_rating WHERE movie_id = $1 AND user_id = $2`, a.RatedMovieID, a.UserID); err != nil {
		b.Fatal(err)
	}

	var tx *sqlx.Tx
	begin, rollback := benchTx(db, &tx)

	runBench(b, []benchCase{
		{name: "GetByID", run: func() error { _, err := r.GetByID(movieID, a.UserID); return err }},
		{name: "GetReviewByMovieID", run: func() error {
			_, err := r.GetReviewByMovieID(strconv.Itoa(a.ReviewedMovieID), a.UserID)
			return err
		}},
		{name: "SeenByUser", run: func() error { _, err := r.SeenByUser(movieID, a.UserID); return err }},
		{name: "IsWatchlisted", run: func() error { _, err := r.IsWatchlisted(movieID, a.UserID); return err }},
		{name: "GetFriends", run: func() error { _, err := r.GetFriends(a.UserID); return err }},
		{name: "GetSeenMovie", run: func() error { _, err := r.GetSeenMovie(a.SeenID, a.UserID); return err }},
		{name: "DeleteSeenMovie", before: begin, after: rollback, run: func() error {
			return r.DeleteSeenMovie(tx, a.SeenID, a.UserID)
		}},
		{name: "UpdateSeenMovie", run: func() error {
			return r.UpdateSeenMovie(a.SeenID, a.UserID, watch.Date, watch.Friends)
		}},
		{name: "DeleteRating", before: begin, after: rollback, run: func() error {
			return r.DeleteRating(tx, a.RatedMovieID, a.UserID)
		}},
		{name: "AddRating", before: begin, after: rollback, run: func() error {
			return r.AddRating(tx, a.UserID, a.RatedMovieID, rating)
		}},
		// Leaves the Rating as it was, apart from updated_at
		{name: "UpdateRating", run: func() error {
			return r.UpdateRating(a.UserID, a.RatedMovieID, strconv.Itoa(rating))
		}},
		{name: "GetUserWatchProviders", run: func() error { _, err := r.GetUserWatchProviders(a.UserID); return err }},
		{name: "DeleteNowPlayingDirect",
			before: func(b *testing.B) {
				benchExec(b, db, `INSERT INTO now_playing (movie_id, user_id, position) VALUES ($1, $2, 0.5)`, a.UnplayedMovieID, a.UserID)
			},
			run: func() error { return r.DeleteNowPlayingDirect(a.UserID, a.UnplayedMovieID) },
		},
		{name: "UpdateNowPlaying",
			run: func() error { return r.UpdateNowPlaying(a.UnplayedMovieID, 0.5, a.UserID) },
			after: func(b *testing.B) {
				benchExec(b, db, `DELETE FROM now_playing WHERE movie_id = $1 AND user_id = $2`, a.UnplayedMovieID, a.UserID)
			},
		},
	})
}

func BenchmarkNowPlaying(b *testing.B) {
	db, a := benchDB(b)
	r := NewNowPlayingRepository(db)

	runBench(b, []benchCase{
		{name: "GetNowPlaying", run: func() error { _, err := r.GetNowPlaying(a.UserID); return err }},
	})
}

func BenchmarkPerson(b *testing.B) {
	db, a := benchDB(b)
	r := NewPersonRepository(db)

	runBench(b, []benchCase{
		{name: "GetPersonByID", run: func() error { _, err := r.GetPersonByID(a.PersonID, a.UserID); return err }},
	})
}

func BenchmarkProductionCompany(b *testing.B) {
	db, a := benchDB(b)
	r := NewProductionCompanyRepository(db)

	runBench(b, []benchCase{
		{name: "GetProductionCompanyMovies", run: func() error {
			_, err := r.GetProductionCompanyMovies(a.CompanyID, a.UserID, 0)
			return err
		}},
		{name: "GetProductionCompanyStats", run: func() error { _, err := r.GetProductionCompanyStats(a.UserID, "All"); return err }},
	})
}

func BenchmarkProductionCountry(b *testing.B) {
	db, a := benchDB(b)
	r := NewProductionCountryRepository(db)

	runBench(b, []benchCase{
		{name: "GetProductionCountryMovies", run: func() error {
			_, err := r.GetProductionCountryMovies(a.CountryID, a.UserID, 0)
			return err
		}},
		{name: "GetProductionCountryStats", run: func() error { _, err := r.GetProductionCountryStats(a.UserID, "All"); return err }},
	})
}

func BenchmarkRatings(b *testing.B) {
	db, a := benchDB(b)
	r := NewRatingsRepository(db)

	runBench(b, []benchCase{
		{name: "GetMoviesByRating", run: func() error { _, err := r.GetMoviesByRating(a.UserID, a.Rating, 0); return err }},
	})
}

func BenchmarkReview(b *testing.B) {
	db, a := benchDB(b)
	r := NewReviewRepository(db)

	runBench(b, []benchCase{
		{name: "InsertReview",
			run: func() error {
				_, err := r.InsertReview("Benchmark review", true, a.UserID, a.UnreviewedMovieID)
				return err
			},
			after: func(b *testing.B) {
				benchExec(b, db, `DELETE FROM review WHERE movie_id = $1 AND user_id = $2`, a.UnreviewedMovieID, a.UserID)
			},
		},
	})
}

func BenchmarkSeries(b *testing.B) {
	db, a := benchDB(b)
	r := NewSeriesRepository(db)

	runBench(b, []benchCase{
		{name: "GetSeriesMovies", run: func() error { _, err := r.GetSeriesMovies(a.SeriesID, a.UserID); return err }},
	})
}

func BenchmarkSettings(b *testing.B) {
	db, a := benchDB(b)
	r := NewSettingsRepository(db)

	providers, err := r.GetWatchProviders(a.UserID)

	if err != nil {
		b.Fatal(err)
	}

	runBench(b, []benchCase{
		{name: "GetWatchProviders", run: func() error { _, err := r.GetWatchProviders(a.UserID); return err }},
		// Writes the user's own providers back
		{name: "UpdateWatchProviders", run: func() error { return r.UpdateWatchProviders(a.UserID, providers) }},
	})
}

func BenchmarkStats(b *testing.B) {
	db, a := benchDB(b)
	r := NewStatsRepository(db)
	// The stats page passes a timestamp for the year-of-date queries
	yearTime := a.WatchYear + "-06-15 12:00:00"

	runBench(b, []benchCase{
		{name: "GetReviewsCount", run: func() error { _, err := r.GetReviewsCount(a.UserID); return err }},
		{name: "GetStatsData", run: func() error { _, err := r.GetStatsData(a.UserID); return err }},
		{name: "GetMostWatchedByJob", run: func() error { _, err := r.GetMostWatchedByJob("cast", a.UserID, "All"); return err }},
		{name: "GetMostWatchedMovies", run: func() error { _, err := r.GetMostWatchedMovies(a.UserID); return err }},
		{name: "GetMoviesByYear", run: func() error { _, err := r.GetMoviesByYear(a.UserID, "All"); return err }},
		{name: "GetRatings", run: func() error { _, err := r.GetRatings(a.UserID); return err }},
		{name: "GetWatchedThisYearByMonth", run: func() error { _, err := r.GetWatchedThisYearByMonth(a.UserID, yearTime); return err }},
		{name: "GetShortestAndLongestMovie", run: func() error { _, err := r.GetShortestAndLongestMovie(a.UserID); return err }},
		{name: "GetTotalWatchedByJobAndYear", run: func() error {
			_, err := r.GetTotalWatchedByJobAndYear(a.UserID, "cast", "All")
			return err
		}},
		{name: "GetWatchedByYear", run: func() error { _, err := r.GetWatchedByYear(a.UserID); return err }},
		{name: "GetWatchedByWeekday", run: func() error { _, err := r.GetWatchedByWeekday(a.UserID, "All"); return err }},
		{name: "GetWilhelmScreamCount", run: func() error { _, err := r.GetWilhelmScreamCount(a.UserID); return err }},
		{name: "GetRatingsThisYear", run: func() error { _, err := r.GetRatingsThisYear(a.UserID, yearTime); return err }},
		{name: "GetMostAwardNominations", run: func() error { _, err := r.GetMostAwardNominations(a.UserID); return err }},
		{name: "GetMostAwardWins", run: func() error { _, err := r.GetMostAwardWins(a.UserID); return err }},
		{name: "GetTopAwardedMovies", run: func() error { _, err := r.GetTopAwardedMovies(a.UserID); return err }},
		{name: "GetRatingsForYear", run: func() error { _, err := r.GetRatingsForYear(a.WatchYear, a.UserID); return err }},
		{name: "GetBestOfTheYear", run: func() error { _, err := r.GetBestOfTheYear(a.UserID, a.WatchYear); return err }},
		{name: "GetWilhelmMovies", run: func() error { _, err := r.GetWilhelmMovies(a.UserID, 0); return err }},
		{name: "GetHighestRankedPersonByJob", run: func() error {
			_, err := r.GetHighestRankedPersonByJob(a.UserID, "cast")
			return err
		}},
	})
}

func BenchmarkWatchlist(b *testing.B) {
	db, a := benchDB(b)
	r := NewWatchlistRepository(db)

	// DeleteFromWatchlist puts the entry back as it was
	var entry struct {
		ID        int       `db:"id"`
		CreatedAt time.Time `db:"created_at"`
	}

	err := db.Get(&entry, `SELECT id, created_at FROM watchlist WHERE movie_id = $1 AND user_id = $2`, a.WatchlistMovieID, a.UserID)

	if err != nil {
		b.Fatal(err)
	}

	runBench(b, []benchCase{
		{name: "GetUnreleasedMovies", run: func() error { _, err := r.GetUnreleasedMovies(a.UserID, "Release date"); return err }},
		{name: "GetReleasedMovies", run: func() error { _, err := r.GetReleasedMovies(a.UserID, "Date added"); return err }},
		{name: "GetTBDMovies", run: func() error { _, err := r.GetTBDMovies(a.UserID); return err }},
		{name: "DeleteFromWatchlist",
			run: func() error { return r.DeleteFromWatchlist(strconv.Itoa(a.WatchlistMovieID), a.UserID) },
			after: func(b *testing.B) {
				benchExec(b, db, `INSERT INTO watchlist (id, movie_id, user_id, created_at) VALUES ($1, $2, $3, $4)`,
					entry.ID, a.WatchlistMovieID, a.UserID, entry.CreatedAt)
			},
		},
	})
}

func BenchmarkYears(b *testing.B) {
	db, a := benchDB(b)
	r := NewYearsRepository(db)

	runBench(b, []benchCase{
		{name: "GetMoviesByYear", run: func() error { _, err := r.GetMoviesByYear(a.UserID, a.ReleaseYear, 0); return err }},
	})
}

// TestBenchmarksCoverUserMethods checks that every Querier method taking a
// userID has a case, named after the method, in the Benchmark named after
// the Querier: StatsQuerier.GetRatings needs "GetRatings" in BenchmarkStats.
func TestBenchmarksCoverUserMethods(t *testing.T) {
	fset := token.NewFileSet()
	files, err := filepath.Glob("*.go")

	if err != nil {
		t.Fatal(err)
	}

	// Benchmark name to the names of the benchCases in its body
	cases := map[string]map[string]bool{}
	// Benchmark name to the methods it needs
	methods := map[string][]string{}

	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, 0)

		if err != nil {
			t.Fatal(err)
		}

		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "Benchmark") && fn.Body != nil {
				names := map[string]bool{}

				// Only a name: field counts, so a string argument like "cast"
				// can't pass for a case
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					kv, ok := n.(*ast.KeyValueExpr)

					if !ok {
						return true
					}

					key, ok := kv.Key.(*ast.Ident)
					lit, isLit := kv.Value.(*ast.BasicLit)

					if ok && isLit && key.Name == "name" && lit.Kind == token.STRING {
						if s, err := strconv.Unquote(lit.Value); err == nil {
							names[s] = true
						}
					}

					return true
				})

				cases[fn.Name.Name] = names
			}
		}

		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		ast.Inspect(f, func(n ast.Node) bool {
			spec, ok := n.(*ast.TypeSpec)

			if !ok || !strings.HasSuffix(spec.Name.Name, "Querier") {
				return true
			}

			iface, ok := spec.Type.(*ast.InterfaceType)

			if !ok {
				return true
			}

			bench := "Benchmark" + strings.TrimSuffix(spec.Name.Name, "Querier")

			for _, m := range iface.Methods.List {
				fn, ok := m.Type.(*ast.FuncType)

				if !ok || len(m.Names) == 0 {
					continue
				}

				for _, param := range fn.Params.List {
					for _, name := range param.Names {
						if name.Name == "userID" {
							methods[bench] = append(methods[bench], m.Names[0].Name)
						}
					}
				}
			}

			return true
		})
	}

	if len(methods) < 10 {
		t.Fatalf("found user methods in only %d Queriers, expected the repositories'", len(methods))
	}

	for bench, names := range methods {
		for _, name := range names {
			if !cases[bench][name] {
				t.Errorf("%s takes a userID but %s has no %q case", name, bench, name)
			}
		}
	}
}
