package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"cmp"
	"strconv"
	"strings"

	"slices"

	"github.com/gofiber/fiber/v2"
)

func HandleGetStats(c *fiber.Ctx) error {
	var stats struct {
		UniqueMovies      int     `db:"unique_movies"`
		SeenWithRewatches int     `db:"seen_with_rewatches"`
		TotalRuntime      int     `db:"total_runtime"`
		TopImdbRating     float64 `db:"top_imdb_rating"`
		TopImdbTitle      string  `db:"top_imdb_title"`
		TopImdbID         string  `db:"top_imdb_id"`
	}

	err := db.Dot.Get(db.Client, &stats, "stats-data")

	if err != nil {
		return err
	}

	return c.Render("stats", fiber.Map{
		"Stats":                 stats,
		"FormattedTotalRuntime": utils.FormatRuntime(stats.TotalRuntime),
	})
}

func HandleGetMostWatchedMovies(c *fiber.Ctx) error {
	var movies []struct {
		Title string `db:"title"`
		ID    string `db:"id"`
		Count int    `db:"count"`
	}

	err := db.Dot.Select(db.Client, &movies, "stats-most-watched-movies")

	if err != nil {
		return err
	}

	return c.Render("partials/stats/most-watched-movies", fiber.Map{
		"Data": movies,
	})
}

func HandleGetMostWatchedByJob(c *fiber.Ctx) error {
	var persons []struct {
		Name  string `db:"name"`
		ID    string `db:"id"`
		Count int    `db:"count"`
	}

	job := c.Params("job")

	err := db.Dot.Select(db.Client, &persons, "stats-most-watched-by-job", job)

	if err != nil {
		return err
	}

	return c.Render("partials/stats/most-watched-person", fiber.Map{
		"Data": persons,
		"Job":  strings.Title(job),
	})
}

type Rating struct {
	Rating int `db:"rating"`
	Count  int `db:"count"`
}

type Bar struct {
	Label     int
	Value     int
	BarHeight int
	BarWidth  int
	BarX      int
	BarY      int
	LabelX    float64
	LabelY    int
	ValueX    float64
	ValueY    int
}

func HandleGetRatings(c *fiber.Ctx) error {
	var ratings []Rating
	var graphData []Bar

	err := db.Dot.Select(db.Client, &ratings, "stats-ratings")

	if err != nil {
		return err
	}

	graphHeight := 200
	graphWidth := 536
	maxCountInRatings := slices.MaxFunc(ratings, func(a, b Rating) int {
		return cmp.Compare(a.Count, b.Count)
	})

	// The data is used for a bar chart, so we need to convert the data
	for i, rating := range ratings {
		var (
			// Calcualte the bar Height
			// Subtract 20 from the maxBarHeight to make room for the text
			barHeight = int(float64(rating.Count) / float64(maxCountInRatings.Count) * float64(graphHeight-20))
			barWidth  = int(float64(graphWidth)/float64(len(ratings))) - 5

			// Space the bars evenly across the graph
			barX = (graphWidth / len(ratings)) * i

			// Subtract the barHeight from the maxBarHeight to position the bar at the bottom.
			// This is because the SVG coordinate system starts at the top left corner.
			barY = graphHeight - barHeight
		)

		// Position centered on the bar. Subtract 3.4 which is half the width of the text.
		charWidth := 7.56 // Uses tabular nums so all characters are the same width
		numberOfCharsInCount := len(strconv.Itoa(rating.Count))
		numberOfCharsInRating := len(strconv.Itoa(rating.Rating))

		halfWidthOfCount := charWidth * float64(numberOfCharsInCount) / 2
		halfWidthOfRating := charWidth * float64(numberOfCharsInRating) / 2

		valueX := float64(barX+(barWidth/2)) - halfWidthOfCount
		labelX := float64(barX+(barWidth/2)) - halfWidthOfRating

		// Subtract 8 to put some space between the text and the bar
		valueY := barY - 8
		// 14 is the height of the text
		labelY := barY + barHeight/2 + 14/2

		// Add the data to the graphData slice
		graphData = append(graphData, Bar{
			Label:     rating.Rating,
			Value:     rating.Count,
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

	return c.Render("partials/stats/ratings", fiber.Map{
		"Data":   graphData,
		"Width":  graphWidth,
		"Height": graphHeight,
	})
}
