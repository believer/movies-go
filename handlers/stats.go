package handlers

import (
	"believer/movies/components"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"cmp"
	"log"
	"strconv"
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

// Handler for /stats.
// Gets most of the necessary data (some is is loaded onload)
func HandleGetStats(c *fiber.Ctx) error {
	var stats types.Stats
	var movies []components.ListItem
	var shortestAndLongest types.Movies
	var wilhelms []int
	var totals []components.ListItem

	userId := c.Locals("UserId").(string)
	now := time.Now()
	year := now.Format("2006")
	currentYear := now.Format("2006-01-02 15:04:05")

	err := db.Dot.Select(db.Client, &movies, "stats-most-watched-movies", userId)

	if err != nil {
		log.Fatalf("Error getting most watched movies: %v", err)
		return err
	}

	err = db.Dot.Select(db.Client, &shortestAndLongest, "shortest-and-longest-movie", userId)

	if err != nil {
		log.Fatalf("Error getting longest and shortest movie: %v", err)
		return err
	}

	err = db.Dot.Select(db.Client, &totals, "total-watched-by-job-and-year", userId, "cast", "All")

	if err != nil {
		return err
	}

	err = db.Dot.Get(db.Client, &stats, "stats-data", userId)

	if err != nil {
		log.Fatalf("Error getting stats data: %v", err)
		return err
	}

	cast, err := getPersonsByJob("cast", userId)

	if err != nil {
		log.Fatalf("Error getting cast: %v", err)
		return err
	}

	ratings, err := getGraphWithQuery("stats-ratings", userId)

	if err != nil {
		log.Fatalf("Error getting ratings: %v", err)
		return err
	}

	yearRatings, err := getGraphByYearWithQuery("stats-ratings-this-year", userId, currentYear)

	if err != nil {
		log.Fatalf("Error getting ratings this year: %v", err)
		return err
	}

	watchedByYear, err := getGraphWithQuery("stats-watched-by-year", userId)

	if err != nil {
		log.Fatalf("Error getting watched by year: %v", err)
		return err
	}

	seenThisYearByMonth, err := getGraphByYearWithQuery("stats-watched-this-year-by-month", userId, currentYear)

	if err != nil {
		log.Fatalf("Error getting watched this year by month: %v", err)
		return err
	}

	moviesByYear, err := getGraphWithQuery("stats-movies-by-year", userId)

	if err != nil {
		log.Fatalf("Error getting movies by year: %v", err)
		return err
	}

	err = db.Dot.Select(db.Client, &wilhelms, "wilhelm-screams", userId)

	if err != nil {
		log.Fatalf("Error getting wilhelm scream: %v", err)
		return err
	}

	var bestOfTheYear types.Movie
	err = db.Dot.Get(db.Client, &bestOfTheYear, "stats-best-of-the-year", userId)

	if err != nil {
		bestOfTheYear = types.Movie{ID: 0}
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

	return utils.TemplRender(c, views.Stats(
		views.StatsProps{
			BestOfTheYear:           bestOfTheYear,
			BestYear:                bestYear,
			FormattedTotalRuntime:   utils.FormatRuntime(stats.TotalRuntime),
			MostWatchedCast:         cast,
			MostWatchedMovies:       movies,
			MoviesByYear:            moviesByYear,
			Ratings:                 ratings,
			SeenThisYear:            seenThisYearByMonth,
			Stats:                   stats,
			TotalCast:               totalCast,
			WatchedByYear:           watchedByYear,
			Year:                    year,
			YearRatings:             yearRatings,
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

	err := db.Dot.Select(db.Client, &genres, "stats-genres", userId)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.MostWatchedGenres(genres))
}
