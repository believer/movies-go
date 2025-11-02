package handlers

import (
	"believer/movies/components/graph"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"cmp"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"slices"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func executeQuery(queryType string, target any, query string, args ...any) func() error {
	return func() error {
		switch queryType {
		case "select":
			return db.Client.Select(target, query, args...)
		case "get":
			return db.Client.Get(target, query, args...)
		default:
			return fmt.Errorf("unknown query type: %s", queryType)
		}
	}
}

type QueryTask struct {
	queryFunc func() error
	desc      string
}

var (
	dataQuery = `
SELECT
    COUNT(DISTINCT movie_id) AS unique_movies,
    COUNT(movie_id) seen_with_rewatches,
    COALESCE(SUM(m.runtime), 0) AS total_runtime
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
WHERE
    user_id = $1
	`

	mostWatchedQuery = `
SELECT
    COUNT(*) AS count,
    m.title AS name,
    m.id
FROM
    seen AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
WHERE
    user_id = $1
GROUP BY
    m.id
ORDER BY
    count DESC
LIMIT 20
`

	ratingsQuery = `
SELECT
    COUNT(*) AS value,
    rating AS label
FROM
    rating
WHERE
    user_id = $1
GROUP BY
    rating
ORDER BY
    rating
	`

	ratingsThisYearQuery = `
WITH rating_series AS (
    SELECT
        generate_series(1, 10) AS rating_value
)
SELECT
    rs.rating_value AS label,
    COUNT(
        CASE WHEN s.movie_id IS NOT NULL THEN
            r.movie_id
        ELSE
            NULL
        END) AS value
FROM
    rating_series rs
    LEFT JOIN rating r ON r.rating = rs.rating_value
        AND r.user_id = $1
    LEFT JOIN seen s ON s.movie_id = r.movie_id
        AND s.user_id = $1
        AND EXTRACT(YEAR FROM s.date) = EXTRACT(YEAR FROM $2::date)
GROUP BY
    rs.rating_value
ORDER BY
    rs.rating_value
`

	mostWatchedByJobQuery = `
SELECT
    COUNT(*) AS count,
    p.name,
    p.id
FROM ( SELECT DISTINCT ON (movie_id)
        movie_id
    FROM
        seen
    WHERE
        user_id = $2
        AND ($3 = 'All'
            OR EXTRACT(YEAR FROM date) = $3::int)) AS s
    INNER JOIN movie_person AS mp ON mp.movie_id = s.movie_id
    INNER JOIN person AS p ON p.id = mp.person_id
WHERE
    mp.job = $1
GROUP BY
    p.id
ORDER BY
    count DESC,
    name ASC
LIMIT 10
`

	totalWatchedByJobAndYearQuery = `
SELECT
    COUNT(*) AS count
FROM
    seen s
    INNER JOIN movie_person mp ON mp.movie_id = s.movie_id
WHERE
    user_id = $1
    AND mp.job = $2
    AND ($3 = 'All'
        OR EXTRACT(YEAR FROM date) = $3::int)
`

	watchedByYearQuery = `
SELECT
    EXTRACT(YEAR FROM date) AS label,
    COUNT(*) AS value
FROM
    seen
WHERE
    user_id = $1
    -- 2011 is where all the data that I hadn't tracked
    -- before I started ended up. So, there's a bunch of
    -- movies that year.
    AND EXTRACT(YEAR FROM date) > 2011
GROUP BY
    label
ORDER BY
    label
`

	watchedThisYearByMonth = `
WITH months (
    month
) AS (
    SELECT
        generate_series(DATE_TRUNC('year', $2::date), DATE_TRUNC('year', $2::date) + INTERVAL '1 year' - INTERVAL '1 day', INTERVAL '1 month'))
SELECT
    TO_CHAR(months.month, 'Mon') AS label,
    COALESCE(count(seen.id), 0) AS value
FROM
    months
    LEFT JOIN seen ON DATE_TRUNC('month', seen.date) = months.month
        AND user_id = $1
WHERE
    EXTRACT(YEAR FROM seen.date) = EXTRACT(YEAR FROM $2::date)
    OR seen.date IS NULL
GROUP BY
    months.month
ORDER BY
    months.month
`

	bestOfYearQuery = `
WITH max_rating AS (
    SELECT
        s.movie_id,
        s.user_id,
        r.rating
    FROM
        seen s
        INNER JOIN rating r ON r.movie_id = s.movie_id
            AND r.user_id = $1
    WHERE
        s.user_id = $1
        AND date >= make_date($2, 1, 1)
        AND date < make_date($2 + 1, 1, 1) -- Seen in the given year
    GROUP BY
        s.id,
        r.rating
    HAVING
        COUNT(*) = 1 -- Seen exactly once in the given year
        AND s.movie_id NOT IN (
            SELECT
                movie_id
            FROM
                seen
            WHERE
                user_id = $1
                AND date < make_date($2, 1, 1) -- Seen before the given year
                OR date >= make_date($2 + 1, 1, 1) -- Seen after the given year
))
SELECT
    m.title AS "name",
    m.id AS "id",
    mr.rating AS "count"
FROM
    max_rating mr
    INNER JOIN movie m ON m.id = mr.movie_id
WHERE
    rating = (
        SELECT
            max(rating)
        FROM
            max_rating)
`

	moviesByYearQuery = `
SELECT
    EXTRACT(YEAR FROM release_date) AS label,
    COUNT(*) AS value
FROM ( SELECT DISTINCT
        movie_id
    FROM
        seen
    WHERE
        user_id = $1) AS s
    INNER JOIN movie AS m ON m.id = s.movie_id
GROUP BY
    label
ORDER BY
    label DESC
`

	shortestLongestQuery = `
(
    SELECT
        m.id,
        m.title,
        m.runtime
    FROM
        movie m
        JOIN seen s ON m.id = s.movie_id
    WHERE
        s.user_id = $1
    ORDER BY
        m.runtime ASC
    LIMIT 1)
UNION ALL (
    SELECT
        m.id,
        m.title,
        m.runtime
    FROM
        movie m
        JOIN seen s ON m.id = s.movie_id
    WHERE
        s.user_id = $1
    ORDER BY
        m.runtime DESC
    LIMIT 1)
`

	highestRankedByJobQuery = `
WITH person_ratings AS (
    SELECT
        p.name,
        p.id,
        COUNT(*) AS appearances,
        SUM(r.rating) AS total_rating
    FROM
        rating AS r
        INNER JOIN movie_person AS mp ON mp.movie_id = r.movie_id
            AND mp.job = $2
        INNER JOIN person AS p ON mp.person_id = p.id
    WHERE
        r.user_id = $1
    GROUP BY
        p.id
)
SELECT
    id,
    name,
    total_rating,
    (total_rating::float / appearances) * LOG(appearances) AS weighted_average_rating,
    appearances
FROM
    person_ratings
ORDER BY
    weighted_average_rating DESC
LIMIT 10
`

	wilhelmQuery = `
SELECT
    count(*)
FROM
    seen s
    INNER JOIN movie m ON m.id = s.movie_id
WHERE
    user_id = $1
    AND m.wilhelm = TRUE
`

	reviewsQuery = `
SELECT
    count(*)
FROM
    review
WHERE
    user_id = $1
`
	mostAwardWinsQuery = `
WITH all_persons AS (
    SELECT DISTINCT ON (mp.person_id)
        mp.person_id
    FROM
        seen s
        INNER JOIN movie_person mp ON mp.movie_id = s.movie_id
    WHERE
        s.user_id = $1
)
SELECT
    count(*) FILTER (WHERE winner = TRUE) AS COUNT,
    a.person,
    a.person_id
FROM
    all_persons ap
    INNER JOIN award a ON ap.person_id = a.person_id
GROUP BY
    a.person_id,
    person
HAVING
    count(*) FILTER (WHERE winner = TRUE) > 0
ORDER BY
    COUNT DESC
LIMIT 1
`
	mostAwardNominationsQuery = `
WITH all_persons AS (
    SELECT DISTINCT ON (mp.person_id)
        mp.person_id
    FROM
        seen s
        INNER JOIN movie_person mp ON mp.movie_id = s.movie_id
    WHERE
        s.user_id = $1
)
SELECT
    count(*) AS COUNT,
    a.person,
    a.person_id
FROM
    all_persons ap
    INNER JOIN award a ON ap.person_id = a.person_id
GROUP BY
    a.person_id,
    person
HAVING
    count(*) > 0
ORDER BY
    COUNT DESC
LIMIT 1
`

	topAwardedQuery = `
WITH movie_awards AS (
    SELECT
        m.id,
        m.title,
        COUNT(DISTINCT a.name) AS award_count
    FROM
        seen s
        INNER JOIN movie m ON m.id = s.movie_id
        INNER JOIN award a ON a.imdb_id = m.imdb_id
    WHERE
        s.user_id = $1
        AND a.winner = TRUE
    GROUP BY
        m.id,
        m.title
)
SELECT
    *
FROM
    movie_awards
WHERE
    award_count = (
        SELECT
            MAX(award_count)
        FROM
            movie_awards)
`
)

// Handler for /stats.
// Gets most of the necessary data (some is is loaded onload)
func GetStats(c *fiber.Ctx) error {
	var stats types.Stats
	var shortestAndLongest types.Movies
	var wilhelms []int
	var movies, totals, cast []types.ListItem
	var ratings, yearRatings, watchedByYear, seenThisYearByMonth, moviesByYear []graph.GraphData
	var awardWins types.AwardPersonStat
	var awardNominations types.AwardPersonStat
	var mostAwardedMovies []types.AwardMovieStat
	var reviews int

	userId := c.Locals("UserId").(string)
	now := time.Now()
	year := now.Format("2006")
	currentYear := now.Format("2006-01-02 15:04:05")

	var wg sync.WaitGroup

	queries := []QueryTask{
		{executeQuery("get", &reviews, reviewsQuery, userId), "stats-reviews"},
		{executeQuery("get", &stats, dataQuery, userId), "stats-data"},
		{executeQuery("select", &cast, mostWatchedByJobQuery, "cast", userId, "All"), "stats-most-watched-by-job"},
		{executeQuery("select", &movies, mostWatchedQuery, userId), "stats-most-watched-movies"},
		{executeQuery("select", &moviesByYear, moviesByYearQuery, userId), "stats-movies-by-year"},
		{executeQuery("select", &ratings, ratingsQuery, userId), "stats-ratings"},
		{executeQuery("select", &seenThisYearByMonth, watchedThisYearByMonth, userId, currentYear), "stats-watched-this-year-by-month"},
		{executeQuery("select", &shortestAndLongest, shortestLongestQuery, userId), "shortest-and-longest-movie"},
		{executeQuery("select", &totals, totalWatchedByJobAndYearQuery, userId, "cast", "All"), "total-watched-by-job-and-year"},
		{executeQuery("select", &watchedByYear, watchedByYearQuery, userId), "stats-watched-by-year"},
		{executeQuery("select", &wilhelms, wilhelmQuery, userId), "wilhelm-screams"},
		{executeQuery("select", &yearRatings, ratingsThisYearQuery, userId, currentYear), "stats-ratings-this-year"},
		{executeQuery("get", &awardNominations, mostAwardNominationsQuery, userId), "stats-most-award-nominations"},
		{executeQuery("get", &awardWins, mostAwardWinsQuery, userId), "stats-most-award-wins"},
		{executeQuery("select", &mostAwardedMovies, topAwardedQuery, userId), "stats-top-awarded-movies"},
	}

	errChan := make(chan error, len(queries))

	for _, q := range queries {
		wg.Add(1)
		go func(queryFunc func() error, desc string) {
			defer wg.Done()
			if err := queryFunc(); err != nil {
				log.Printf("Error getting %s: %v", desc, err)
				errChan <- err
			}
		}(q.queryFunc, q.desc)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	totalCast := 0

	if len(totals) > 0 {
		totalCast = totals[0].Count
	}

	// Process all graph data in parallel
	errChan = make(chan error, 5)
	var ratingsBars, yearBars, watchedByYearBar, seenThisYearByMonthBars []graph.Bar

	wg.Add(4)

	go func() {
		defer wg.Done()
		var err error
		ratingsBars, err = constructGraphFromData(ratings)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		yearBars, err = constructGraphFromData(yearRatings)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		watchedByYearBar, err = constructGraphFromData(watchedByYear)
		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		var err error
		seenThisYearByMonthBars, err = constructGraphFromData(seenThisYearByMonth)
		if err != nil {
			errChan <- err
		}
	}()

	// Close the error channel once all goroutines are done
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	bestYear := ""
	maxYear := 0

	for _, m := range moviesByYear {
		if m.Value > maxYear {
			maxYear = m.Value
			bestYear = m.Label
		}
	}

	return utils.Render(c, views.Stats(
		views.StatsProps{
			AwardNominations:        awardNominations,
			AwardWins:               awardWins,
			BestYear:                bestYear,
			FormattedTotalRuntime:   utils.FormatRuntime(stats.TotalRuntime),
			MostAwardedMovies:       mostAwardedMovies,
			MostWatchedCast:         cast,
			MostWatchedMovies:       movies,
			MoviesByYear:            moviesByYear,
			Ratings:                 ratingsBars,
			Reviews:                 reviews,
			SeenThisYear:            seenThisYearByMonthBars,
			Stats:                   stats,
			TotalCast:               utils.Formatter().Sprintf("%d", totalCast),
			WatchedByYear:           watchedByYearBar,
			Year:                    year,
			YearRatings:             yearBars,
			Years:                   availableYears(),
			ShortestAndLongestMovie: shortestAndLongest,
			WilhelmScreams:          wilhelms[0],
		}))
}

func GetMostWatchedByJob(c *fiber.Ctx) error {
	var persons []types.ListItem
	var totals []types.ListItem

	job := c.Params("job")
	year := c.Query("year", "All")
	userId := c.Locals("UserId")
	years := availableYears()
	years = append([]string{"All"}, years...)

	err := db.Client.Select(&persons, mostWatchedByJobQuery, job, userId, year)

	if err != nil {
		return err
	}

	err = db.Client.Select(&totals, totalWatchedByJobAndYearQuery, userId, job, year)

	if err != nil {
		return err
	}

	totalJob := 0

	if len(totals) > 0 {
		totalJob = totals[0].Count
	}
	return utils.Render(c, views.MostWatchedPerson(
		views.MostWatchedPersonProps{
			Data:  persons,
			Job:   job,
			Title: cases.Title(language.English).String(job),
			Total: utils.Formatter().Sprintf("%d", totalJob),
			Year:  year,
			Years: years,
		}))
}

func GetHighestRankedPersonByJob(c *fiber.Ctx) error {
	var persons []types.HighestRated

	job := c.Query("job", "cast")
	userId := c.Locals("UserId")
	title := "Highest ranked " + strings.ToLower(job)

	err := db.Client.Select(&persons, highestRankedByJobQuery, userId, strings.ToLower(job))

	if err != nil {
		return err
	}

	jobs := []string{"Cast", "Composer", "Director", "Producer", "Writer"}

	return utils.Render(c, views.HighestRating(
		views.HighestRatingProps{
			Data:  persons,
			Job:   job,
			Jobs:  jobs,
			Title: title,
		}))
}

func getGraphByYearWithQuery(query string, userId string, year string) ([]graph.Bar, error) {
	var data []graph.GraphData

	err := db.Client.Select(&data, query, userId, year)

	if err != nil {
		return nil, err
	}

	return constructGraphFromData(data)
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func constructGraphFromData(data []graph.GraphData) ([]graph.Bar, error) {
	var graphData []graph.Bar

	graphHeight := 200
	graphWidth := 536
	maxCount := 0

	if len(data) > 0 {
		maxCount = slices.MaxFunc(data, func(a, b graph.GraphData) int {
			return cmp.Compare(a.Value, b.Value)
		}).Value
	}

	// The data is used for a bar chart, so we need to convert the data
	for i, row := range data {
		var (
			elementsInGraph = graphWidth / len(data)
			// Calcualte the bar Height
			// Subtract 46 from the graph height to make room for the labels
			barHeight = clamp(int(float64(row.Value)/float64(maxCount)*float64(graphHeight-46)), 2, graphHeight-46)
			barWidth  = clamp(int(float64(graphWidth)/float64(len(data)))-5, 2, 100)

			// Space the bars evenly across the graph
			barX = elementsInGraph*i + 1
			barY = graphHeight - barHeight - 26
		)

		// Position centered on the bar. Subtract 3.4 which is half the width of the text.
		charWidth := 8.67 // Uses tabular nums so all characters are the same width
		numberOfCharsInCount := len(strconv.Itoa(row.Value))
		numberOfCharsInLabel := len(row.Label)

		halfWidthOfCount := charWidth * float64(numberOfCharsInCount) / 2
		halfWidthOfLabel := charWidth * float64(numberOfCharsInLabel) / 2

		valueX := float64(barX+(barWidth/2)) - halfWidthOfCount
		labelX := float64(barX+(barWidth/2)) - halfWidthOfLabel

		// Subtract 8 to put some space between the text and the bar
		valueY := barY - 8
		// 16,5 is the height of the text
		labelY := float64(barY) + float64(barHeight) + 20

		// Add the data to the graphData slice
		graphData = append(graphData, graph.Bar{
			Label:     row.Label,
			Value:     row.Value,
			BarHeight: barHeight,
			BarWidth:  barWidth,
			BarX:      barX,
			BarY:      barY,
			ValueX:    valueX,
			ValueY:    valueY,
			LabelX:    labelX,
			LabelY:    labelY,
		})
	}

	return graphData, nil
}

func GetRatingsByYear(c *fiber.Ctx) error {
	userId := c.Locals("UserId").(string)
	year := c.Query("year")
	currentYear := time.Now().Format("2006")
	yearTime, err := pgSelectedYear(year)

	if err != nil {
		return err
	}

	yearRatings, err := getGraphByYearWithQuery("stats-ratings-this-year", userId, yearTime)

	if err != nil {
		return err
	}

	title := "Ratings " + year

	if currentYear == year {
		title = "Ratings this year"
	}

	return utils.Render(c, graph.WithYear(
		graph.WithYearProps{
			Props: graph.Props{
				Bars:  yearRatings,
				Title: title,
			},
			SelectedYear: year,
			Years:        availableYears(),
			Route:        "/stats/ratings",
		}))
}

func GetThisYearByMonth(c *fiber.Ctx) error {
	userId := c.Locals("UserId").(string)
	year := c.Query("year")
	currentYear := time.Now().Format("2006")
	yearTime, err := pgSelectedYear(year)

	if err != nil {
		return err
	}

	yearRatings, err := getGraphByYearWithQuery(watchedThisYearByMonth, userId, yearTime)

	if err != nil {
		return err
	}

	title := "Seen " + year + " by month"

	if currentYear == year {
		title = "Seen this year by month"
	}

	return utils.Render(c, graph.WithYear(
		graph.WithYearProps{
			Props: graph.Props{
				Bars:  yearRatings,
				Title: title,
			},
			SelectedYear: year,
			Years:        availableYears(),
			Route:        "/stats/by-month",
		}))
}

func pgSelectedYear(year string) (string, error) {
	parsedYear, err := strconv.Atoi(year)

	if err != nil {
		return "", err
	}

	return time.Date(parsedYear, time.September, 10, 0, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05"), nil
}

func availableYears() []string {
	// First year with "real" data
	// 2011 is used as a catch all for anything before I had the database
	endYear := 2012
	currentYear := time.Now().Year()

	years := make([]string, 0)

	for year := currentYear; year >= endYear; year-- {
		y := strconv.Itoa(year)
		years = append(years, y)
	}

	return years
}

func GetBestOfTheYear(c *fiber.Ctx) error {
	var movies []types.ListItem

	userId := c.Locals("UserId")
	currentYear := time.Now().Format("2006")
	year := c.Query("year", currentYear)
	years := availableYears()

	err := db.Client.Select(&movies, bestOfYearQuery, userId, year)

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  movies,
		Title: "Best of the Year",
		Route: "/stats/best-of-the-year",
		Root:  "movie",
		Year:  year,
		Years: years,
	}))
}
