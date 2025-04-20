package app

import (
	"believer/movies/db"
	"believer/movies/router"
	"believer/movies/types"
	"believer/movies/utils"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func readOscarsFile() {
	f, err := os.Open("oscars.csv")

	if err != nil {
		panic(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()

	if err != nil {
		panic(err)
	}

	// Gather indexes to headers for easier mapping
	fields := make(map[string]int)
	for i, name := range records[0] {
		fields[name] = i
	}

	tx := db.Client.MustBegin()

	for _, r := range records {
		year := r[fields["Year"]]
		category := r[fields["Category"]]
		imdbId := r[fields["FilmId"]]
		detail := r[fields["Detail"]]
		winner := r[fields["Winner"]] == "TRUE"

		// Use to update one movie
		// if r[fields["Film"]] != "Gladiator" {
		// 	continue
		// }

		// Use to update one job
		if category != "Film Editing" {
			continue
		}

		fmt.Println(year)

		// We can only add where movie exists in database, otherwise
		// we get a violation on the foreign key to the movie table
		var movie types.Movie
		err := tx.Get(&movie, `SELECT id FROM movie WHERE imdb_id = $1`, imdbId)

		if err != nil {
			continue
		}

		switch category {
		case "Actor in a Leading Role",
			"Actor in a Supporting Role",
			"Actress in a Leading Role",
			"Actress in a Supporting Role",
			"Cinematography",
			"Directing",
			"Film Editing",
			"Music (Original Score)",
			"Writing (Adapted Screenplay)",
			"Writing (Original Screenplay)":
			nominees := strings.Split(r[fields["Nominees"]], ", ")

			if len(nominees) == 0 {
				continue
			}

			for _, n := range nominees {
				var person types.Person

				err = tx.Get(&person, `SELECT p."name", p.id FROM movie m
    INNER JOIN movie_person mp ON mp.movie_id = m.id
    INNER JOIN person p ON p.id = mp.person_id
  WHERE m.imdb_id = $1
  AND p."name" ILIKE '%' || $2 || '%'`, imdbId, n)

				if err != nil {
					continue
				}

				_, err = tx.Exec(`INSERT INTO award (name, imdb_id, winner, year, person, person_id) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (imdb_id, name, year, person, detail) DO UPDATE SET winner = excluded.winner,name = excluded.name, person_id = excluded.person_id`, category, imdbId, winner, year, n, person.ID)

				if err != nil {
					fmt.Printf("Person Err %s %s %t %s %s %d\n", category, imdbId, winner, year, n, person.ID)
					panic(err)
				}

				fmt.Printf("Person %s\n", n)
			}
		case "Music (Original Song)":
			_, err = tx.Exec(`INSERT INTO award (name, imdb_id, winner, year, detail) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (imdb_id, name, year, person, detail) DO UPDATE SET winner = excluded.winner`, category, imdbId, winner, year, detail)

			if err != nil {
				fmt.Printf("Music Err %s %s %t %s %s\n", category, imdbId, winner, year, detail)
				panic(err)
			}

			fmt.Printf("Music %s\n", detail)
		default:
			_, err = tx.Exec(`INSERT INTO award (name, imdb_id, winner, year) VALUES ($1, $2, $3, $4) ON CONFLICT (imdb_id, name, year, person, detail) DO UPDATE SET winner = excluded.winner`, category, imdbId, winner, year)

			if err != nil {
				fmt.Printf("Other Err %s %s %t %s\n", category, imdbId, winner, year)
				panic(err)
			}

			fmt.Printf("Other %s %s\n", category, imdbId)
		}
	}

	err = tx.Commit()

	if err != nil {
		err = tx.Rollback()

		panic(err)
	}
}

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
