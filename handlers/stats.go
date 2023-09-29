package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"cmp"
	"strconv"

	"slices"

	"github.com/gofiber/fiber/v2"
)

func HandleGetStats(c *fiber.Ctx) error {
	var stats struct {
		UniqueMovies      int `db:"unique_movies"`
		SeenWithRewatches int `db:"seen_with_rewatches"`
		TotalRuntime      int `db:"total_runtime"`
	}

	err := db.Client.Get(&stats, `
SELECT
	COUNT(DISTINCT movie_id) AS unique_movies,
	COUNT(movie_id) seen_with_rewatches,
  SUM(m.runtime) AS total_runtime
FROM
	seen AS s
INNER JOIN movie as m ON m.id = s.movie_id 
WHERE
	user_id = 1;
    `)

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

type Rating struct {
	Rating int `db:"rating"`
	Count  int `db:"count"`
}

func HandleGetRatings(c *fiber.Ctx) error {
	var ratings []Rating

	err := db.Dot.Select(db.Client, &ratings, "stats-ratings")

	if err != nil {
		return err
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

	var graphData []Bar

	maxBarHeight := 200
	graphWidth := 536
	maxCountInRatings := slices.MaxFunc(ratings, func(a, b Rating) int {
		return cmp.Compare(a.Count, b.Count)
	})

	// The data is used for a bar chart, so we need to convert the data
	for i, rating := range ratings {
		// Calcualte the bar Height
		// Subtract 20 from the maxBarHeight to make room for the text
		barHeight := int(float64(rating.Count) / float64(maxCountInRatings.Count) * float64(maxBarHeight-20))
		barWidth := int(float64(graphWidth)/float64(len(ratings))) - 5

		// Space the bars evenly across the graph
		barX := (graphWidth / len(ratings)) * i

		// Subtract the barHeight from the maxBarHeight to position the bar at the bottom.
		// This is because the SVG coordinate system starts at the top left corner.
		barY := maxBarHeight - barHeight

		// Position centered on the bar. Subtract 3.4 which is half the width of the text.
		charWidth := 7.56 // Uses tabular nums so all characters are the same width
		numberOfCharsInCount := len(strconv.Itoa(rating.Count))
		halfWidthOfCount := charWidth * float64(numberOfCharsInCount) / 2
		valueX := float64(barX+(barWidth/2)) - halfWidthOfCount
		labelX := float64(barX+(barWidth/2)) - halfWidthOfCount

		// Subtract 8 to put some space between the text and the bar
		valueY := barY - 8
		labelY := maxBarHeight + 15

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
		"Height": maxBarHeight,
	})
}
