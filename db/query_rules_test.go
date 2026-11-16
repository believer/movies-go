package db

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// queryRule is a CODING_STANDARDS.md rule that a query's text shows it
// breaks, so it's checked here rather than in review.
type queryRule struct {
	// standard names the CODING_STANDARDS.md entry
	standard string
	// check returns why the query breaks the rule, or "" when it doesn't
	check func(query string) string
	// known lists the owners (see sqlLiteral) that broke the rule before it
	// was checked. Fix one and remove it here; the test fails on stale entries.
	known []string
}

var queryRules = []queryRule{
	{
		standard: "Latest rating",
		check:    readsRatingTable,
	},
	{
		standard: "Watched movies",
		check:    distinctWatchedMovies,
	},
	{
		standard: "Deterministic order",
		check:    cutoffWithoutUniqueKey,
		known: []string{
			"FeedRepository.GetFeedMovies",
			"FeedRepository.SearchPersons",
			"StatsRepository.GetMostWatchedByJob",
			"StatsRepository.GetMostWatchedMovies",
			"StatsRepository.GetShortestAndLongestMovie",
			"StatsRepository.GetMostAwardNominations",
			"StatsRepository.GetMostAwardWins",
			"genreStatsQuery",
			"languageStatsQuery",
			"listProductionCompaniesQuery",
			"productionCompanyStatsQuery",
			"productionCountryStatsQuery",
		},
	},
}

// TestQueryRules checks every SQL literal in this package against
// queryRules. It reads the source only, so it runs with -short.
func TestQueryRules(t *testing.T) {
	literals := sqlLiterals(t)

	for _, rule := range queryRules {
		t.Run(rule.standard, func(t *testing.T) {
			broken := map[string]bool{}

			for _, lit := range literals {
				reason := rule.check(lit.Query)

				if reason == "" {
					continue
				}

				broken[lit.Owner] = true

				if !slices.Contains(rule.known, lit.Owner) {
					t.Errorf("%s (%s): %s. See %q in CODING_STANDARDS.md.", lit.Pos, lit.Owner, reason, rule.standard)
				}
			}

			for _, owner := range rule.known {
				if !broken[owner] {
					t.Errorf("%s follows %q now: remove it from the rule's known list", owner, rule.standard)
				}
			}
		})
	}
}

var (
	readStart       = regexp.MustCompile(`(?i)^[\s(]*(SELECT|WITH)\b`)
	ratingTableRead = regexp.MustCompile(`(?i)\b(FROM|JOIN)\s+(public\.)?rating\b`)
)

// readsRatingTable flags a read of the rating table, which returns every
// Rating rather than the latest. Writes start with INSERT, UPDATE or DELETE.
func readsRatingTable(query string) string {
	if !readStart.MatchString(query) || !ratingTableRead.MatchString(query) {
		return ""
	}

	return "reads the rating table, read latest_rating instead"
}

// watchedMovieCopy matches the distinct movies a user watched written out by
// hand: a DISTINCT over seen filtered on user_id alone, as a subquery or CTE.
// Filtering on more (a year) is a different set, which the view can't give.
var watchedMovieCopy = regexp.MustCompile(`(?is)SELECT\s+DISTINCT(\s+ON\s*\(\s*movie_id\s*\))?\s+movie_id\s*(,\s*id\s*)?FROM\s+(public\.)?seen(\s+\w+)?\s+WHERE\s+user_id\s*=\s*(\$\d+|\w+)\s*(\)|ORDER\s+BY\b)`)

// distinctWatchedMovies flags a hand-written copy of watched_movie.
func distinctWatchedMovies(query string) string {
	if !watchedMovieCopy.MatchString(query) {
		return ""
	}

	return "builds the user's distinct watched movies from seen, read watched_movie instead"
}

var (
	orderBy   = regexp.MustCompile(`\border\s+by\b`)
	clauseEnd = regexp.MustCompile(`^\b(limit|offset|union|fetch)\b`)
	uniqueKey = regexp.MustCompile(`^\(?"?([a-z_][a-z0-9_]*\.)?([a-z0-9_]*_)?id"?\)?(\s+(asc|desc))?(\s+nulls\s+(first|last))?$`)
)

// cutoffWithoutUniqueKey flags an ORDER BY followed by LIMIT or OFFSET whose
// last key isn't an id, so ties can fall either side of the cutoff. It
// can't tell which other keys are unique, so it asks for an id-named key.
func cutoffWithoutUniqueKey(query string) string {
	q := maskSQL(strings.ToLower(query))
	depth := make([]int, len(q)+1)

	for i, c := range q {
		depth[i+1] = depth[i]

		switch c {
		case '(':
			depth[i+1]++
		case ')':
			depth[i+1]--
		}
	}

	for _, m := range orderBy.FindAllStringIndex(q, -1) {
		d := depth[m[0]]
		start, end := m[1], len(q)
		cutoff := false

		for i := start; i < len(q); i++ {
			if depth[i+1] < d || q[i] == ';' {
				end = i
				break
			}

			if depth[i] == d {
				if kw := clauseEnd.FindString(q[i:]); kw != "" && (i == 0 || !isWordByte(q[i-1])) {
					end = i
					cutoff = kw == "limit" || kw == "offset"
					break
				}
			}
		}

		if !cutoff {
			continue
		}

		key := lastKey(q[start:end], depth[start:end], d)

		if !uniqueKey.MatchString(key) {
			return fmt.Sprintf("ORDER BY ends on %q before a LIMIT or OFFSET, end it on a unique key such as m.id", key)
		}
	}

	return ""
}

// lastKey returns the last comma-separated key of an ORDER BY clause,
// splitting only at the clause's own depth.
func lastKey(clause string, depth []int, d int) string {
	last := 0

	for i := range clause {
		if clause[i] == ',' && depth[i] == d {
			last = i + 1
		}
	}

	return strings.Join(strings.Fields(clause[last:]), " ")
}

// maskSQL blanks out comments and string literals, keeping offsets, so
// their contents don't count as keywords, parentheses or commas.
func maskSQL(q string) string {
	b := []byte(q)

	for i := 0; i < len(b); i++ {
		switch {
		case b[i] == '-' && i+1 < len(b) && b[i+1] == '-':
			for ; i < len(b) && b[i] != '\n'; i++ {
				b[i] = ' '
			}
		case b[i] == '\'':
			b[i] = ' '

			for i++; i < len(b) && b[i] != '\''; i++ {
				b[i] = ' '
			}

			if i < len(b) {
				b[i] = ' '
			}
		}
	}

	return string(b)
}

func isWordByte(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9'
}

func TestCutoffWithoutUniqueKey(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		broken bool
	}{
		{"id last", `SELECT id FROM movie ORDER BY release_date DESC, m.id OFFSET $1 LIMIT 50`, false},
		{"id-named key last", `SELECT movie_id FROM seen GROUP BY movie_id ORDER BY COUNT(*) DESC, movie_id DESC LIMIT 10`, false},
		{"no cutoff", `SELECT id FROM movie ORDER BY release_date DESC`, false},
		{"non-unique key before LIMIT", `SELECT id FROM movie ORDER BY release_date DESC LIMIT 10`, true},
		{"non-unique key before OFFSET", `SELECT id FROM movie ORDER BY release_date DESC OFFSET $1`, true},
		{"cutoff in a subquery", `(SELECT id FROM movie ORDER BY runtime ASC LIMIT 1) UNION ALL (SELECT id FROM movie ORDER BY runtime DESC, id LIMIT 1)`, true},
		{"ORDER BY inside an aggregate", `SELECT array_agg(title ORDER BY title) FROM movie GROUP BY id ORDER BY id LIMIT 5`, false},
		{"ORDER BY inside a window", `SELECT ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at) FROM rating ORDER BY id LIMIT 5`, false},
		{"keywords in strings and comments", `SELECT 'limit' AS l FROM movie ORDER BY title -- limit 1
`, false},
		{"commas in the last key", `SELECT id FROM movie ORDER BY id, COALESCE(title, name) LIMIT 1`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := cutoffWithoutUniqueKey(tt.query)

			if (reason != "") != tt.broken {
				t.Errorf("got %q, want broken = %v", reason, tt.broken)
			}
		})
	}
}

func TestReadsRatingTable(t *testing.T) {
	tests := []struct {
		query  string
		broken bool
	}{
		{`SELECT rating FROM rating WHERE user_id = $1`, true},
		{`SELECT m.id FROM movie m JOIN public.rating r ON r.movie_id = m.id`, true},
		{`SELECT rating FROM latest_rating WHERE user_id = $1`, false},
		{`SELECT m.rating_count FROM movie m`, false},
		{`DELETE FROM rating WHERE id = $1`, false},
		{`INSERT INTO rating (user_id, movie_id, rating) VALUES ($1, $2, $3)`, false},
	}

	for _, tt := range tests {
		if (readsRatingTable(tt.query) != "") != tt.broken {
			t.Errorf("%q: want broken = %v", tt.query, tt.broken)
		}
	}
}

func TestDistinctWatchedMovies(t *testing.T) {
	tests := []struct {
		query  string
		broken bool
	}{
		{`LEFT JOIN ( SELECT DISTINCT ON (movie_id)
            movie_id,
            id
        FROM
            public.seen
        WHERE
            user_id = $2
        ORDER BY
            movie_id,
            id) AS s ON m.id = s.movie_id`, true},
		{`WITH seen_once AS (SELECT DISTINCT movie_id FROM seen s WHERE user_id = $1) SELECT COUNT(*) FROM seen_once`, true},
		{`FROM ( SELECT DISTINCT ON (movie_id) movie_id FROM seen WHERE user_id = $1
        AND ($2 = 'All' OR EXTRACT(YEAR FROM date) = $2::int)) AS s`, false},
		{`LEFT JOIN watched_movie AS w ON w.movie_id = m.id AND w.user_id = $1`, false},
		{`SELECT s.date FROM seen s WHERE s.user_id = $1 ORDER BY s.date DESC, s.id`, false},
	}

	for _, tt := range tests {
		if (distinctWatchedMovies(tt.query) != "") != tt.broken {
			t.Errorf("%q: want broken = %v", tt.query, tt.broken)
		}
	}
}
