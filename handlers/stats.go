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

	err := db.Dot.Select(db.Client, &persons, "stats-most-watched-by-job", job, userId)

	if err != nil {
		return nil, err
	}

	return persons, nil
}

func HandleGetStats(c *fiber.Ctx) error {
	var stats types.Stats
	var movies []components.ListItem

	userId := c.Locals("UserId").(string)
	now := time.Now()
	currentYear := now.Format("2006-01-02 15:04:05")

	err := db.Dot.Select(db.Client, &movies, "stats-most-watched-movies", userId)

	if err != nil {
		log.Fatalf("Error getting most watched movies: %v", err)
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

	year := now.Format("2006")

	return utils.TemplRender(c, views.Stats(
		views.StatsProps{
			Stats:                 stats,
			FormattedTotalRuntime: utils.FormatRuntime(stats.TotalRuntime),
			MostWatchedCast:       cast,
			WatchedByYear:         watchedByYear,
			Ratings:               ratings,
			YearRatings:           yearRatings,
			MostWatchedMovies:     movies,
			SeenThisYear:          seenThisYearByMonth,
			BestOfTheYear:         bestOfTheYear,
			MoviesByYear:          moviesByYear,
			BestYear:              bestYear,
			Year:                  year,
			Years:                 availableYears(),
		}))
}

func HandleGetMostWatchedByJob(c *fiber.Ctx) error {
	job := c.Params("job")
	persons, err := getPersonsByJob(job, c.Locals("UserId").(string))

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.MostWatchedPerson(persons,
		cases.Title(language.English).String(job),
	))
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
	maxCount := slices.MaxFunc(data, func(a, b types.GraphData) int {
		return cmp.Compare(a.Value, b.Value)
	})

	// The data is used for a bar chart, so we need to convert the data
	for i, row := range data {
		var (
			elementsInGraph = graphWidth / len(data)
			// Calcualte the bar Height
			// Subtract 46 from the graph height to make room for the labels
			barHeight = clamp(int(float64(row.Value)/float64(maxCount.Value)*float64(graphHeight-46)), 2, graphHeight-46)
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
	parsedYear, err := strconv.Atoi(year)
	currentYear := time.Now().Year()

	if err != nil {
		return err
	}

	yearTime := time.Date(parsedYear, time.September, 10, 0, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")
	yearRatings, err := getGraphByYearWithQuery("stats-ratings-this-year", userId, yearTime)

	if err != nil {
		return err
	}

	title := "Ratings " + year

	if currentYear == parsedYear {
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
	parsedYear, err := strconv.Atoi(year)
	currentYear := time.Now().Year()

	if err != nil {
		return err
	}

	yearTime := time.Date(parsedYear, time.September, 10, 0, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05")
	yearRatings, err := getGraphByYearWithQuery("stats-watched-this-year-by-month", userId, yearTime)

	if err != nil {
		return err
	}

	title := "Seen " + year + " by month"

	if currentYear == parsedYear {
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

func availableYears() []int {
	// First year with "real" data
	// 2011 is used as a catch all for anything before I had the database
	startYear := 2012
	currentYear := time.Now().Year()

	years := make([]int, 0)

	for year := startYear; year <= currentYear; year++ {
		years = append(years, year)
	}

	return years
}
