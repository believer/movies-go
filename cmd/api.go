package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"

	h "believer/movies/handlers"
	repo "believer/movies/internal/adapters/sqlc"
	"believer/movies/internal/movies"
	"believer/movies/utils"
	"believer/movies/views"
)

type api struct {
	config config
	db     *pgx.Conn
}

type dbConfig struct {
	dsn string
}

type config struct {
	addr string
	db   dbConfig
}

func redirectToHome(c *fiber.Ctx) error {
	return c.Redirect("/")
}

func NotFoundMiddleware(c *fiber.Ctx) error {
	c.Status(fiber.StatusNotFound)
	return utils.Render(c, views.NotFound())
}

func (api *api) mount() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	appEnv := os.Getenv("APP_ENV")

	// Recover middleware recovers from panics anywhere in
	// the chain and handles the control to the centralized ErrorHandler.
	app.Use(recover.New())

	// Logger middleware will log the HTTP requests.
	app.Use(logger.New())

	// Add CSRF token
	store := session.New()
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "cookie:csrf_",
		CookieName:     "csrf_",
		CookieSameSite: "Lax",
		CookieSecure:   appEnv != "development",
		CookieHTTPOnly: true,
		Session:        store,
		SessionKey:     "fiber.csrf.token",
	}))

	// Setup locals
	app.Use(func(c *fiber.Ctx) error {
		secret := os.Getenv("ADMIN_SECRET")
		tokenString := c.Cookies("token")

		// Set me as default user
		userId := "1"

		// Parse the JWT token if it exists
		// and set the user ID in the locals
		if tokenString != "" {
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				// Validate the signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				return []byte(secret), nil
			})

			if err != nil {
				log.Fatal(err)
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				userId = claims["id"].(string)
			}
		}

		c.Locals("AppEnv", appEnv)
		c.Locals("UserId", userId)
		c.Locals("IsAuthenticated", utils.IsAuthenticated(c))

		return c.Next()
	})

	// Serve static files
	app.Static("/robots.txt", "./public/robots.txt")
	app.Static("/public", "./public", fiber.Static{
		MaxAge: 31_536_000, // 1 year
	})

	app.Get("/health", h.GetHealth)
	app.Get("/", h.GetFeed)

	// Auth
	// --------------------------
	app.Get("/login", h.GetLogin)
	app.Post("/login", h.Login)
	app.Post("/logout", h.Logout)

	// Watchlist
	// --------------------------
	watchlistGroup := app.Group("/watchlist")

	watchlistGroup.Get("/", h.GetWatchlist)
	watchlistGroup.Get("/movies", h.GetWatchlistMovies)
	watchlistGroup.Get("/unreleased-movies", h.GetWatchlistUnreleasedMovies)
	watchlistGroup.Delete("/:id", h.DeleteFromWatchlist)

	// Movies
	// --------------------------

	movieService := movies.NewService(repo.New(api.db))
	movieHandler := movies.NewHandler(movieService)

	movieGroup := app.Group("/movie")

	movieGroup.Get("/", redirectToHome)
	movieGroup.Get("/imdb", h.GetByImdbId)
	movieGroup.Get("/search", h.HandleSearch)
	movieGroup.Get("/new", h.GetMovieNew)
	movieGroup.Get("/new/series", h.GetMovieNewSeries)
	// TODO: Should really be post to /movie
	movieGroup.Post("/new", h.PostMovieNew)

	movieGroup.Get("/:id", movieHandler.Movie)
	movieGroup.Patch("/:id", h.UpdateMovieByID)
	movieGroup.Get("/:imdbId/awards", h.GetMovieAwards)
	movieGroup.Get("/:id/seen/others", h.GetMovieOthersSeenByID)
	movieGroup.Post("/:id/seen", h.CreateSeenMovie)
	movieGroup.Delete("/:id/seen/:seenId", h.DeleteSeenMovie)
	movieGroup.Get("/:id/seen/:seenId/edit", h.GetSeenMovie)
	movieGroup.Put("/:id/seen/:seenId/edit", h.UpdateSeenMovie)

	movieGroup.Get("/:id/rating", h.GetRating)
	movieGroup.Get("/:id/rating/edit", h.GetEditRating)
	movieGroup.Post("/:id/rating", h.PostRating)
	movieGroup.Put("/:id/rating", h.UpdateRating)
	movieGroup.Delete("/:id/rating", h.DeleteRating)

	movieGroup.Get("/:id/watch-providers", h.WatchProviders)

	// Review
	// --------------------------
	reviewGroup := app.Group("/review")

	reviewGroup.Get("/:id/edit", h.EditMovieReview)
	reviewGroup.Post("/:id/update", h.UpdateMovieReview)

	// Year
	// --------------------------
	yearGroup := app.Group("/year")

	yearGroup.Get("/", redirectToHome)
	yearGroup.Get("/:year", h.GetMoviesByYear)

	// Person
	// --------------------------
	personGroup := app.Group("/person")

	personGroup.Get("/", redirectToHome)
	personGroup.Get("/:id", h.GetPersonByID)

	// Genre
	// --------------------------
	genreGroup := app.Group("/genre")

	genreGroup.Get("/", redirectToHome)
	genreGroup.Get("/stats", h.GetGenreStats)
	genreGroup.Get("/:id", h.GetGenre)

	// Language
	// --------------------------
	languageGroup := app.Group("/language")

	languageGroup.Get("/", redirectToHome)
	languageGroup.Get("/stats", h.GetLanguageStats)
	languageGroup.Get("/:id", h.GetLanguage)

	// Production companies
	// --------------------------
	productionCompanyGroup := app.Group("/production-company")

	productionCompanyGroup.Get("/", redirectToHome)
	productionCompanyGroup.Get("/stats", h.GetProductionCompanyStats)
	productionCompanyGroup.Get("/:id", h.GetProductionCompany)

	// Production countries
	// --------------------------
	productionCountryGroup := app.Group("/production-country")

	productionCountryGroup.Get("/", redirectToHome)
	productionCountryGroup.Get("/stats", h.GetProductionCountryStats)
	productionCountryGroup.Get("/:id", h.GetProductionCountry)

	// Awards
	// --------------------------
	awardsGroup := app.Group("/awards")

	awardsGroup.Get("/", redirectToHome)
	awardsGroup.Get("/:awards", h.GetMoviesByNumberOfAwards)
	awardsGroup.Get("/year/:year", h.GetAwardsByYear)

	// Rating
	// --------------------------
	ratingGroup := app.Group("/rating")

	ratingGroup.Get("/", redirectToHome)
	ratingGroup.Get("/:rating", h.GetMoviesByRating)

	// Stats
	// --------------------------
	statsGroup := app.Group("/stats")

	statsGroup.Get("/", h.GetStats)
	statsGroup.Get("/ratings", h.GetRatingsByYear)
	statsGroup.Get("/by-month", h.GetThisYearByMonth)
	statsGroup.Get("/highest-ranked-person", h.GetHighestRankedPersonByJob)
	statsGroup.Get("/best-of-the-year", h.GetBestOfTheYear)
	statsGroup.Get("/most-watched-person/:job", h.GetMostWatchedByJob)

	// Series
	// --------------------------
	seriesGroup := app.Group("/series")

	seriesGroup.Get("/:id", h.GetSeries)

	// Settings
	// --------------------------
	settingsGroup := app.Group("/settings")

	settingsGroup.Get("/", h.Settings)
	settingsGroup.Put("/watch-providers", h.SettingsWatchProviders)

	app.Use(NotFoundMiddleware)

	return app
}

func (api *api) run(a *fiber.App) error {
	log.Info("Server started on " + api.config.addr)

	return a.Listen(api.config.addr)
}
