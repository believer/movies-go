package handlers

import (
	"believer/movies/components"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"cmp"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"slices"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func getPersonsByJob(job string, userId string) ([]components.ListItem, error) {
	var persons []components.ListItem

	err := db.Dot.Select(db.Client, &persons, "stats-most-watched-by-job", job, userId, "All")

	if err != nil {
		return nil, err
	}

	return persons, nil
}

func executeQuery(queryType string, target interface{}, query string, args ...interface{}) func() error {
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
func HandleGetStats(c *fiber.Ctx) error {
	var stats types.Stats
	var shortestAndLongest types.Movies
	var wilhelms []int
	var movies, totals, cast []components.ListItem
	var ratings, yearRatings, watchedByYear, seenThisYearByMonth, moviesByYear []types.GraphData

	bestOfTheYear := types.Movie{ID: 0}

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
		{executeQuery("get", &bestOfTheYear, "stats-best-of-the-year", userId), "stats-best-of-the-year"},
		{executeQuery("select", &cast, "stats-most-watched-by-job", "cast", userId, "All"), "stats-most-watched-by-job"},
		{executeQuery("select", &ratings, "stats-ratings", userId), "stats-ratings"},
		{executeQuery("select", &yearRatings, "stats-ratings-this-year", userId, currentYear), "stats-ratings-this-year"},
		{executeQuery("select", &watchedByYear, "stats-watched-by-year", userId), "stats-watched-by-year"},
		{executeQuery("select", &seenThisYearByMonth, "stats-watched-this-year-by-month", userId, currentYear), "stats-watched-this-year-by-month"},
		{executeQuery("select", &moviesByYear, "stats-movies-by-year", userId), "stats-movies-by-year"},
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

	bestYear := ""
	bestYearValue := 0

	for _, year := range moviesByYear {
		if year.Value > bestYearValue {
			bestYear = year.Label
			bestYearValue = year.Value
		}
	}

	totalCast := 0

	if len(totals) > 0 {
		totalCast = totals[0].Count
	}

	// Process all graph data in parallel
	errChan = make(chan error, 5)
	var ratingsBars, yearBars, watchedByYearBar, seenThisYearByMonthBars, moviesByYearBars []types.Bar

	wg.Add(5)

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

	go func() {
		defer wg.Done()
		var err error
		moviesByYearBars, err = constructGraphFromData(moviesByYear)
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

	return utils.TemplRender(c, views.Stats(
		views.StatsProps{
			BestOfTheYear:           bestOfTheYear,
			BestYear:                bestYear,
			FormattedTotalRuntime:   utils.FormatRuntime(stats.TotalRuntime),
			MostWatchedCast:         cast,
			MostWatchedMovies:       movies,
			MoviesByYear:            moviesByYearBars,
			Ratings:                 ratingsBars,
			SeenThisYear:            seenThisYearByMonthBars,
			Stats:                   stats,
			TotalCast:               totalCast,
			WatchedByYear:           watchedByYearBar,
			Year:                    year,
			YearRatings:             yearBars,
			Years:                   availableYears(),
			ShortestAndLongestMovie: shortestAndLongest,
			WilhelmScreams:          wilhelms[0],
		}))
}

func HandleGetMostWatchedByJob(c *fiber.Ctx) error {
	var persons []components.ListItem
	var totals []components.ListItem

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

	return utils.TemplRender(c, components.MostWatchedPerson(
		components.MostWatchedPersonProps{
			Data:  persons,
			Job:   job,
			Title: cases.Title(language.English).String(job),
			Total: totalJob,
			Year:  year,
			Years: years,
		}))
}

func getGraphWithQuery(query string, userId string) ([]types.Bar, error) {
	var data []types.GraphData

	err := db.Dot.Select(db.Client, &data, query, userId)

	if err != nil {
		return nil, err
	}

	return constructGraphFromData(data)
}

func getGraphByYearWithQuery(query string, userId string, year string) ([]types.Bar, error) {
	var data []types.GraphData

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

func constructGraphFromData(data []types.GraphData) ([]types.Bar, error) {
	var graphData []types.Bar

	graphHeight := 200
	graphWidth := 536
	maxCount := 0

	if len(data) > 0 {
		maxCount = slices.MaxFunc(data, func(a, b types.GraphData) int {
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
		graphData = append(graphData, types.Bar{
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

func HandleGetRatingsByYear(c *fiber.Ctx) error {
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

	return utils.TemplRender(c, components.GraphWithYear(
		components.GraphWithYearProps{
			Bars:         yearRatings,
			Title:        title,
			SelectedYear: year,
			Years:        availableYears(),
			Route:        "/stats/ratings",
		}))
}

func HandleGetThisYearByMonth(c *fiber.Ctx) error {
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

	return utils.TemplRender(c, components.GraphWithYear(
		components.GraphWithYearProps{
			Bars:         yearRatings,
			Title:        title,
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

func HandleGetGenreStats(c *fiber.Ctx) error {
	var genres []components.ListItem

	userId := c.Locals("UserId").(string)
	year := c.Query("year", "All")
	years := append([]string{"All"}, availableYears()...)

	err := db.Dot.Select(db.Client, &genres, "stats-genres", userId, year)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.MostWatchedGenres(components.MostWatchedGenresProps{
		Data:  genres,
		Year:  year,
		Years: years,
	}))
}
