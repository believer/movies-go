package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HandleFeed(c *fiber.Ctx) error {
	var movies types.Movies

	page := c.QueryInt("page", 1)
	searchQuery := c.Query("search")

	if searchQuery != "" {
		err := db.Dot.Select(db.Client, &movies, "feed-search", searchQuery)

		if err != nil {
			return err
		}
	} else {
		err := db.Dot.Select(db.Client, &movies, "feed", (page-1)*20, c.Locals("UserId"))

		if err != nil {
			return err
		}
	}

	feed := views.Feed(
		utils.IsAuthenticated(c),
		movies,
		page+1,
		searchQuery,
	)

	return utils.TemplRender(c, feed)
}

func HandleGetLogin(c *fiber.Ctx) error {
	return utils.TemplRender(c, views.Login())
}

type LoginData struct {
	PasswordHash string `db:"password_hash"`
	ID           string `db:"id"`
}

func HandlePostLogin(c *fiber.Ctx) error {
	var user LoginData

	data := new(struct {
		Password string `form:"password"`
		Username string `form:"username"`
	})

	// Parse the form data
	if err := c.BodyParser(data); err != nil {
		return err
	}

	// Get the password hash of the user from the database
	err := db.Client.Get(&user, "SELECT id, password_hash FROM public.user WHERE username = $1", data.Username)

	if err != nil {
		c.Set("HX-Retarget", "#error")
		c.Set("HX-Reswap", "innerHTML")
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid username or password")
	}

	// Check if the password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.Password))

	if err != nil {
		c.Set("HX-Retarget", "#error")
		c.Set("HX-Reswap", "innerHTML")
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid username or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": user.ID,
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("ADMIN_SECRET")))

	if err != nil {
		err := c.SendStatus(401)

		if err != nil {
			return err
		}

		return c.SendString("Server error")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().AddDate(0, 0, 30),
		HTTPOnly: true,
	})

	return c.Redirect("/", 303)
}

func HandlePostLogout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().AddDate(0, 0, -1),
		HTTPOnly: true,
	})

	c.Set("HX-Redirect", "/")

	return c.SendStatus(200)
}

func HandlePostSignup(c *fiber.Ctx) error {
	data := new(struct {
		Password string `form:"password"`
		Username string `form:"username"`
	})

	// Parse the form data
	if err := c.BodyParser(data); err != nil {
		return err
	}

	if data.Username == "" {
		return c.SendString("Missing username")
	}

	if data.Password == "" {
		return c.SendString("Missing password")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	tx := db.Client.MustBegin()

	tx.MustExec(`INSERT INTO "user" (username, password_hash) VALUES ($1, $2)`, data.Username, string(hash))

	err = tx.Commit()

	if err != nil {
		err = tx.Rollback()

		return err
	}

	return c.SendStatus(fiber.StatusOK)
}
