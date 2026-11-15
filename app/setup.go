package app

import (
	"believer/movies/components/layout"
	"believer/movies/db"
	"believer/movies/router"
	"believer/movies/utils"
	"believer/movies/views"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/golang-jwt/jwt/v5"
)

type Headers struct {
	Authorization string `reqHeader:"Authorization"`
}

// newApp builds the Fiber app. It is separate from SetupAndRunApp so the error
// handler can be tested without a database or a listener.
func newApp() *fiber.App {
	return fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler:          handleError,
	})
}

// handleError logs unexpected errors and sends a generic body. Fiber's default
// handler sends err.Error() to the client, which can leak SQL, internal paths
// or secrets carried in the error. Fiber errors are deliberate responses, so
// they keep their status and message.
func handleError(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		code = fiberErr.Code
		message = fiberErr.Message
	}

	if code >= fiber.StatusInternalServerError {
		slog.Error("request failed", "method", c.Method(), "path", c.OriginalURL(), "error", err)
		message = "Internal Server Error"
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)

	return c.Status(code).SendString(message)
}

// SetupAndRunApp starts the app. cssSources holds the CSS files the
// stylesheet is built from.
func SetupAndRunApp(cssSources fs.FS) error {
	utils.LoadEnv()
	appEnv := os.Getenv("APP_ENV")

	// Build the stylesheet and version its URL by its contents to bust caches
	css, err := layout.BuildStylesheet(cssSources)

	if err != nil {
		return err
	}

	layout.VersionStylesheet(css)

	// Initialize database connection
	err = db.InitializeConnection()

	if err != nil {
		return err
	}

	// Close database connection when the app exits
	defer db.CloseConnection()

	// Setup the app
	app := newApp()

	// Recover middleware recovers from panics anywhere in
	// the chain and handles the control to the centralized ErrorHandler.
	app.Use(recover.New())

	// Logger middleware will log the HTTP requests.
	app.Use(logger.New())

	// Pass app environment to all views
	app.Use(func(c *fiber.Ctx) error {
		secret := os.Getenv("ADMIN_SECRET")
		tokenString := c.Cookies("token")
		c.Locals("IsAuthenticated", false)

		h := new(Headers)

		if err := c.ReqHeaderParser(h); err != nil {
			return err
		}

		if tokenString == "" && h.Authorization != "" {
			_, token, _ := strings.Cut(h.Authorization, " ")
			tokenString = token
		}

		// Set me as default user
		userId := "1"

		// Parse the JWT token if it exists
		// and set the user ID in the locals
		if tokenString != "" {
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				// Validate the signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				return []byte(secret), nil
			})

			if err != nil {
				return err
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				userId = claims["id"].(string)
				c.Locals("IsAuthenticated", true)
			}
		}

		c.Locals("AppEnv", appEnv)
		c.Locals("UserId", userId)

		return c.Next()
	})

	// Serve static files. The stylesheet URL is versioned, so it can be
	// cached for good
	app.Get("/styles.css", func(c *fiber.Ctx) error {
		c.Type("css")
		c.Set(fiber.HeaderCacheControl, "public, max-age=31536000, immutable")

		return c.Send(css)
	})
	app.Static("/robots.txt", "./public/robots.txt")
	app.Static("/public", "./public", fiber.Static{
		MaxAge: 31_536_000, // 1 year
	})

	// Setup routes
	router.SetupRoutes(app)

	app.Use(notFoundMiddleware)

	// Start the app
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))

	return nil
}

func notFoundMiddleware(c *fiber.Ctx) error {
	c.Status(fiber.StatusNotFound)
	return utils.Render(c, views.NotFound())
}
