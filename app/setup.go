package app

import (
	"believer/movies/db"
	"believer/movies/router"
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
		tokenString := c.Cookies("token")

		c.Locals("AppEnv", appEnv)

		if tokenString != "" {
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				// Don't forget to validate the alg is what you expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
				}

				// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
				return []byte(os.Getenv("ADMIN_SECRET")), nil
			})

			if err != nil {
				log.Fatal(err)
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				c.Locals("UserId", claims["id"])
			} else {
				fmt.Println(err)
			}
		}

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
