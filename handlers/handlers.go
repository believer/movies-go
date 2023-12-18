package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
	"believer/movies/views"
	"encoding/base64"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func HandleFeed(c *fiber.Ctx) error {
	var movies types.Movies

	pageQuery := c.Query("page", "1")
	page, err := strconv.Atoi(pageQuery)

	if err != nil {
		page = 1
	}

	err = db.Dot.Select(db.Client, &movies, "feed", (page-1)*20)

	if err != nil {
		panic(err)
	}

	if c.Get("Accept") == "application/json" {
		return c.JSON(movies)
	}

	feed := views.Feed(
		utils.IsAuthenticated(c),
		movies,
		page+1,
	)

	return utils.TemplRender(c, feed)
}

func HandleGetLogin(c *fiber.Ctx) error {
	return utils.TemplRender(c, views.Login(""))
}

func HandlePostLogin(c *fiber.Ctx) error {
	data := new(struct {
		Password string `form:"password"`
		Username string `form:"username"`
	})

	if err := c.BodyParser(data); err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(data.Username + ":" + data.Password))

	if encoded == os.Getenv("ADMIN_SECRET") {
		c.Cookie(&fiber.Cookie{
			Name:     "admin_secret",
			Value:    encoded,
			Expires:  time.Now().AddDate(0, 0, 30),
			HTTPOnly: true,
		})

		return c.Redirect("/", 303)
	}

	return utils.TemplRender(c, views.Login("Invalid username or password"))
}

func HandlePostLogout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "admin_secret",
		Value:    "",
		Expires:  time.Now().AddDate(0, 0, -1),
		HTTPOnly: true,
	})

	c.Set("HX-Redirect", "/")

	return c.SendStatus(200)
}
