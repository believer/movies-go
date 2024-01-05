package app

import (
	"believer/movies/db"
	"believer/movies/router"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func SetupAndRunApp() error {
	// Load environment variables
	godotenv.Load()

	// Initialize database connection
	err := db.InitializeConnection()

	if err != nil {
		return err
	}

	// Close database connection when the app exits
	defer db.CloseConnection()

	// Setup the app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	// Setup middleware
	// Recover middleware recovers from panics anywhere in the chain and handles the control to the centralized ErrorHandler.
	app.Use(recover.New())
	// Logger middleware will log the HTTP requests.
	app.Use(logger.New())

	// Pass app environment to all views
	app.Use(func(c *fiber.Ctx) error {
		appEnv := os.Getenv("APP_ENV")

		c.Locals("AppEnv", appEnv)

		return c.Next()
	})

	// Serve static files
	app.Static("/public", "./public", fiber.Static{
		MaxAge:   86400, // 1 day
		Compress: true,
	})

	// Setup routes
	router.SetupRoutes(app)

	// Start the app
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	app.Listen(":" + port)

	return nil
}
