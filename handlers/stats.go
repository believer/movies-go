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
			return db.Dot.Select(db.Client, target, query, args...)
		case "get":
			return db.Dot.Get(db.Client, target, query, args...)
		default:
			return fmt.Errorf("unknown query type: %s", queryType)
		}
	}
}

type QueryTask struct {
	queryFunc func() error
	desc      string
}

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
		{executeQuery("select", &movies, "stats-most-watched-movies", userId), "stats-most-watched-movies"},
		{executeQuery("select", &shortestAndLongest, "shortest-and-longest-movie", userId), "shortest-and-longest-movie"},
		{executeQuery("select", &totals, "total-watched-by-job-and-year", userId, "cast", "All"), "total-watched-by-job-and-year"},
		{executeQuery("get", &stats, "stats-data", userId), "stats-data"},
		{executeQuery("select", &wilhelms, "wilhelm-screams", userId), "wilhelm-screams"},
		{executeQuery("select", &cast, "stats-most-watched-by-job", "cast", userId, "All"), "stats-most-watched-by-job"},
		{executeQuery("select", &ratings, "stats-ratings", userId), "stats-ratings"},
		{executeQuery("select", &yearRatings, "stats-ratings-this-year", userId, currentYear), "stats-ratings-this-year"},
		{executeQuery("select", &watchedByYear, "stats-watched-by-year", userId), "stats-watched-by-year"},
		{executeQuery("select", &seenThisYearByMonth, "stats-watched-this-year-by-month", userId, currentYear), "stats-watched-this-year-by-month"},
		{executeQuery("select", &moviesByYear, "stats-movies-by-year", userId), "stats-movies-by-year"},
		{executeQuery("get", &awardWins, "stats-most-award-wins", userId), "stats-most-award-wins"},
		{executeQuery("get", &awardNominations, "stats-most-award-nominations", userId), "stats-most-award-nominations"},
		{executeQuery("select", &mostAwardedMovies, "stats-top-awarded-movies", userId), "stats-top-awarded-movies"},
		{executeQuery("get", &reviews, "stats-reviews", userId), "stats-reviews"},
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

	err := db.Dot.Select(db.Client, &persons, "stats-most-watched-by-job", job, userId, year)

	if err != nil {
		return err
	}

	err = db.Dot.Select(db.Client, &totals, "total-watched-by-job-and-year", userId, job, year)

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

	err := db.Dot.Select(db.Client, &persons, "stats-highest-ranked-persons-by-job", userId, strings.ToLower(job))

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

	err := db.Dot.Select(db.Client, &data, query, userId, year)

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

	yearRatings, err := getGraphByYearWithQuery("stats-watched-this-year-by-month", userId, yearTime)

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

	err := db.Dot.Select(db.Client, &movies, "best-of-the-year", userId, year)

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
