package handlers

import (
	"believer/movies/components"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"cmp"
	"strconv"

	"slices"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func getPersonsByJob(job string) ([]components.ListItem, error) {
	var persons []components.ListItem

	err := db.Dot.Select(db.Client, &persons, "stats-most-watched-by-job", job)

	if err != nil {
		return nil, err
	}

	return persons, nil
}

func HandleGetStats(c *fiber.Ctx) error {
	var stats types.Stats
	var movies []components.ListItem

	err := db.Dot.Select(db.Client, &movies, "stats-most-watched-movies")

	if err != nil {
		return err
	}

	err = db.Dot.Get(db.Client, &stats, "stats-data")

	if err != nil {
		return err
	}

	cast, err := getPersonsByJob("cast")

	if err != nil {
		return err
	}

	ratings, err := getGraphWithQuery("stats-ratings")

	if err != nil {
		return err
	}

	watchedByYear, err := getGraphWithQuery("stats-watched-by-year")

	if err != nil {
		return err
	}

	seenThisYearByMonth, err := getGraphWithQuery("stats-watched-this-year-by-month")

	if err != nil {
		return err
	}

	return utils.TemplRender(c, views.Stats(
		stats,
		utils.FormatRuntime(stats.TotalRuntime),
		cast,
		watchedByYear,
		ratings,
		movies,
		seenThisYearByMonth,
	))
}

func HandleGetMostWatchedByJob(c *fiber.Ctx) error {
	job := c.Params("job")
	persons, err := getPersonsByJob(job)

	if err != nil {
		return err
	}

	return utils.TemplRender(c, components.MostWatchedPerson(persons,
		cases.Title(language.English).String(job),
	))
}

func getGraphWithQuery(query string) ([]types.Bar, error) {
	var data []types.GraphData

	err := db.Dot.Select(db.Client, &data, query)

	if err != nil {
		return nil, err
	}

	return constructGraphFromData(data)
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
			// Calcualte the bar Height
			// Subtract 20 from the maxBarHeight to make room for the text
			barHeight = int(float64(row.Value) / float64(maxCount.Value) * float64(graphHeight-40))
			barWidth  = int(float64(graphWidth)/float64(len(data))) - 5

			// Space the bars evenly across the graph
			barX = (graphWidth / len(data)) * i
			barY = graphHeight - barHeight - 20
		)

		// Position centered on the bar. Subtract 3.4 which is half the width of the text.
		charWidth := 8.67 // Uses tabular nums so all characters are the same width
		numberOfCharsInCount := len(strconv.Itoa(row.Value))
		numberOfCharsInRating := len(strconv.Itoa(row.Label))

		halfWidthOfCount := charWidth * float64(numberOfCharsInCount) / 2
		halfWidthOfRating := charWidth * float64(numberOfCharsInRating) / 2

		valueX := float64(barX+(barWidth/2)) - halfWidthOfCount
		labelX := float64(barX+(barWidth/2)) - halfWidthOfRating

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
