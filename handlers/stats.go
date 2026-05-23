package handlers

import (
	"believer/movies/components/graph"
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type StatsHandler struct {
	repo db.StatsQuerier
}

func NewStatsHandler(repo db.StatsQuerier) *StatsHandler {
	return &StatsHandler{repo}
}

// Handler for /stats.
func (h *StatsHandler) GetStats(c *fiber.Ctx) error {
	var reviews int
	var stats types.Stats
	var cast []types.ListItem
	var movies []types.ListItem
	var moviesByYear []graph.GraphData
	var ratings []graph.GraphData
	var seenThisYearByMonth []graph.GraphData
	var shortestAndLongest types.Movies
	var totals []types.ListItem
	var watchedByYear []graph.GraphData
	var watchedByWeekday []graph.GraphData
	var wilhelms []int
	var yearRatings []graph.GraphData
	var awardNominations types.AwardPersonStat
	var awardWins types.AwardPersonStat
	var mostAwardedMovies []types.AwardMovieStat

	var errReviews, errStats, errCast, errMovies, errMoviesByYear, errRatings, errSeenThisYearByMonth, errShortestAndLongest, errTotals, errWatchedByYear, errWatchedByWeekday, errWilhelms, errYearRatings, errAwardNominations, errAwardWins, errMostAwardedMovies error

	userId := c.Locals("UserId").(string)
	now := time.Now()
	year := now.Format("2006")
	currentYear := now.Format("2006-01-02 15:04:05")

	var wg sync.WaitGroup
	wg.Add(16)

	go func() {
		defer wg.Done()
		reviews, errReviews = h.repo.GetReviewsCount(userId)
	}()

	go func() {
		defer wg.Done()
		stats, errStats = h.repo.GetStatsData(userId)
	}()

	go func() {
		defer wg.Done()
		cast, errCast = h.repo.GetMostWatchedByJob("cast", userId, "All")
	}()

	go func() {
		defer wg.Done()
		movies, errMovies = h.repo.GetMostWatchedMovies(userId)
	}()

	go func() {
		defer wg.Done()
		moviesByYear, errMoviesByYear = h.repo.GetMoviesByYear(userId, "All")
	}()

	go func() {
		defer wg.Done()
		ratings, errRatings = h.repo.GetRatings(userId)
	}()

	go func() {
		defer wg.Done()
		seenThisYearByMonth, errSeenThisYearByMonth = h.repo.GetWatchedThisYearByMonth(userId, currentYear)
	}()

	go func() {
		defer wg.Done()
		shortestAndLongest, errShortestAndLongest = h.repo.GetShortestAndLongestMovie(userId)
	}()

	go func() {
		defer wg.Done()
		totals, errTotals = h.repo.GetTotalWatchedByJobAndYear(userId, "cast", "All")
	}()

	go func() {
		defer wg.Done()
		watchedByYear, errWatchedByYear = h.repo.GetWatchedByYear(userId)
	}()

	go func() {
		defer wg.Done()
		watchedByWeekday, errWatchedByWeekday = h.repo.GetWatchedByWeekday(userId, "All")
	}()

	go func() {
		defer wg.Done()
		wilhelms, errWilhelms = h.repo.GetWilhelmScreamCount(userId)
	}()

	go func() {
		defer wg.Done()
		yearRatings, errYearRatings = h.repo.GetRatingsThisYear(userId, currentYear)
	}()

	go func() {
		defer wg.Done()
		awardNominations, errAwardNominations = h.repo.GetMostAwardNominations(userId)
	}()

	go func() {
		defer wg.Done()
		awardWins, errAwardWins = h.repo.GetMostAwardWins(userId)
	}()

	go func() {
		defer wg.Done()
		mostAwardedMovies, errMostAwardedMovies = h.repo.GetTopAwardedMovies(userId)
	}()

	wg.Wait()

	// Check errors
	for _, err := range []error{errReviews, errStats, errCast, errMovies, errMoviesByYear, errRatings, errSeenThisYearByMonth, errShortestAndLongest, errTotals, errWatchedByYear, errWatchedByWeekday, errWilhelms, errYearRatings, errAwardNominations, errAwardWins, errMostAwardedMovies} {
		if err != nil {
			return err
		}
	}

	totalCast := 0

	if len(totals) > 0 {
		totalCast = totals[0].Count
	}

	// Process all graph data in parallel
	var ratingsBars, yearBars, watchedByYearBar, watchedByWeekdayBar, seenThisYearByMonthBars []graph.Bar
	var errChan = make(chan error, 5)

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
		watchedByWeekdayBar, err = constructGraphFromData(watchedByWeekday)
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
		wg.Wait()
		close(errChan)
	}()

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
			WatchedByWeekday:        watchedByWeekdayBar,
			Year:                    year,
			YearRatings:             yearBars,
			Years:                   availableYears(),
			ShortestAndLongestMovie: shortestAndLongest,
			WilhelmScreams:          wilhelms[0],
		}))
}

func (h *StatsHandler) GetMostWatchedByJob(c *fiber.Ctx) error {
	var persons []types.ListItem
	var totals []types.ListItem

	job := c.Params("job")
	year := c.Query("year", "All")
	userId := c.Locals("UserId").(string)
	years := availableYears()
	years = append([]string{"All"}, years...)

	persons, err := h.repo.GetMostWatchedByJob(job, userId, year)
	if err != nil {
		return err
	}

	totals, err = h.repo.GetTotalWatchedByJobAndYear(userId, job, year)
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

func (h *StatsHandler) GetHighestRankedPersonByJob(c *fiber.Ctx) error {
	var persons []types.HighestRated

	job := c.Query("job", "cast")
	userId := c.Locals("UserId").(string)
	title := "Highest ranked " + strings.ToLower(job)

	persons, err := h.repo.GetHighestRankedPersonByJob(userId, strings.ToLower(job))

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

func (h *StatsHandler) getGraphByYearWithQuery(userID string, year string, queryFunc func(string, string) ([]graph.GraphData, error)) ([]graph.Bar, error) {
	data, err := queryFunc(userID, year)

	if err != nil {
		return nil, err
	}

	return constructGraphFromData(data)
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

func (h *StatsHandler) GetRatingsByYear(c *fiber.Ctx) error {
	userId := c.Locals("UserId").(string)
	year := c.Query("year")
	currentYear := time.Now().Format("2006")
	yearTime, err := pgSelectedYear(year)

	if err != nil {
		return err
	}

	yearRatings, err := h.getGraphByYearWithQuery(userId, yearTime, h.repo.GetRatingsThisYear)

	if err != nil {
		return err
	}

	title := "Ratings " + year

	if currentYear == year {
		title = "Ratings this year"
	}

	return utils.Render(c, graph.WithYear(
		graph.WithYearProps{
			SectionHref: fmt.Sprintf("/stats/ratings/%s", year),
			BarHref:     "/rating",
			Props: graph.Props{
				Bars:  yearRatings,
				Title: title,
			},
			SelectedYear: year,
			Years:        availableYears(),
			Route:        "/stats/ratings",
		}))
}

func (h *StatsHandler) GetRatingsForYear(c *fiber.Ctx) error {
	userId := c.Locals("UserId").(string)
	year := c.Params("year")

	movies, err := h.repo.GetRatingsForYear(year, userId)

	if err != nil {
		return err
	}

	sorted := make([]types.Movies, 10)

	for _, m := range movies {
		key := m.Rating.Int64 - 1
		sorted[key] = append(sorted[key], m)
	}

	slices.Reverse(sorted)

	return utils.Render(c, views.Ratings(views.RatingsProps{
		Movies: sorted,
		Title:  fmt.Sprintf("Ratings in %s", year),
	}))
}

func (h *StatsHandler) GetThisYearByMonth(c *fiber.Ctx) error {
	userId := c.Locals("UserId").(string)
	year := c.Query("year")
	currentYear := time.Now().Format("2006")
	yearTime, err := pgSelectedYear(year)

	if err != nil {
		return err
	}

	yearRatings, err := h.getGraphByYearWithQuery(userId, yearTime, h.repo.GetWatchedThisYearByMonth)

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

func (h *StatsHandler) GetThisYearByWeekday(c *fiber.Ctx) error {
	userId := c.Locals("UserId").(string)
	year := c.Query("year")
	yearTime, err := pgSelectedYear(year)

	if err != nil {
		return err
	}

	yearRatings, err := h.getGraphByYearWithQuery(userId, yearTime, h.repo.GetWatchedByWeekday)

	if err != nil {
		return err
	}

	title := "Seen " + year + " by weekday"

	if year == "All" {
		title = "Seen by weekday"
	}

	return utils.Render(c, graph.WithYear(
		graph.WithYearProps{
			Props: graph.Props{
				Bars:  yearRatings,
				Title: title,
			},
			SelectedYear: year,
			Years:        append([]string{"All"}, availableYears()...),
			Route:        "/stats/by-weekday",
		}))
}

func pgSelectedYear(year string) (string, error) {
	if year == "All" {
		return "All", nil
	}

	parsedYear, err := strconv.Atoi(year)

	if err != nil {
		return "", err
	}

	return time.Date(parsedYear, time.September, 10, 0, 0, 0, 0, time.UTC).Format("2006-01-02 15:04:05"), nil
}

func availableYears() []string {
	endYear := 2012
	currentYear := time.Now().Year()

	years := make([]string, 0)

	for year := currentYear; year >= endYear; year-- {
		y := strconv.Itoa(year)
		years = append(years, y)
	}

	return years
}

func (h *StatsHandler) GetBestOfTheYear(c *fiber.Ctx) error {
	userId := c.Locals("UserId").(string)
	currentYear := time.Now().Format("2006")
	year := c.Query("year", currentYear)
	years := availableYears()

	movies, err := h.repo.GetBestOfTheYear(userId, year)

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

func (h *StatsHandler) GetWilhelmScream(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	userId := c.Locals("UserId").(string)

	movies, err := h.repo.GetWilhelmMovies(userId, (page-1)*50)

	if err != nil {
		return err
	}

	return utils.Render(c, views.ListView(views.ListViewProps{
		EmptyState: "No movies with the Wilhelm scream",
		Name:       "Wilhelm screams",
		NextPage:   fmt.Sprintf("/stats/wilhelm-scream?page=%d", page+1),
		Movies:     movies,
	}))
}

func (h *StatsHandler) GetSeenWith(c *fiber.Ctx) error {
	items, err := h.repo.GetSeenWith()

	if err != nil {
		return err
	}

	return utils.Render(c, views.StatsSection(views.StatsSectionProps{
		Data:  items,
		Title: "Seen with",
	}))
}

func (h *StatsHandler) GetMoviesByYearStat(c *fiber.Ctx) error {
	userId := c.Locals("UserId").(string)
	year := c.Query("year", "All")

	moviesByYear, err := h.repo.GetMoviesByYear(userId, year)

	if err != nil {
		return err
	}

	bestYear := ""
	maxYear := 0

	for _, m := range moviesByYear {
		if m.Value > maxYear {
			maxYear = m.Value
			bestYear = m.Label
		}
	}

	return utils.Render(c, views.MoviesByYear(views.MoviesByYearProps{
		BestYear: bestYear,
		Movies:   moviesByYear,
		Year:     year,
		Years:    append([]string{"All"}, availableYears()...),
	}))
}
