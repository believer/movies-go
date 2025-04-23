package app

import (
	"believer/movies/db"
	"believer/movies/router"
	"believer/movies/utils"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func SetupAndRunApp() error {
	// Load environment variables
	err := godotenv.Load()

	if err != nil && os.Getenv("APP_ENV") == "development" {
		return err
	}

	// Initialize database connection
	err = db.InitializeConnection()

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

	// Recover middleware recovers from panics anywhere in
	// the chain and handles the control to the centralized ErrorHandler.
	app.Use(recover.New())

	// Logger middleware will log the HTTP requests.
	app.Use(logger.New())

	// Pass app environment to all views
	app.Use(func(c *fiber.Ctx) error {
		secret := os.Getenv("ADMIN_SECRET")
		appEnv := os.Getenv("APP_ENV")
		tokenString := c.Cookies("token")

		// Set me as default user
		userId := "1"

		// Parse the JWT token if it exists
		// and set the user ID in the locals
		if tokenString != "" {
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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

	// Setup routes
	router.SetupRoutes(app)

	// Start the app
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))

	return nil
}
