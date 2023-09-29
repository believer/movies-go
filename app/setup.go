package app

import (
	"believer/movies/db"
	"believer/movies/router"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
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

	// Setup templates
	engine := html.New("./views", ".html")

	// Add custom functions to the template engine
	engine.AddFunc("StringsJoin", strings.Join)

	// Simple add function
	engine.AddFunc(
		"add", func(x, y int) int {
			return x + y
		},
	)

	// Add possibility to send a map to the template
	engine.AddFunc("map", func(pairs ...any) (map[string]any, error) {
		if len(pairs)%2 != 0 {
			return nil, errors.New("misaligned map")
		}

		m := make(map[string]any, len(pairs)/2)

		for i := 0; i < len(pairs); i += 2 {
			key, ok := pairs[i].(string)

			if !ok {
				return nil, fmt.Errorf("cannot use type %T as map key", pairs[i])
			}
			m[key] = pairs[i+1]
		}
		return m, nil
	})

	// Setup the app
	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "layouts/main",
	})

	// Setup middleware
	// Recover middleware recovers from panics anywhere in the chain and handles the control to the centralized ErrorHandler.
	app.Use(recover.New())
	// Logger middleware will log the HTTP requests.
	app.Use(logger.New())

	// Serve static files
	app.Static("/public", "./public")

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
