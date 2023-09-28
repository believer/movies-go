package handlers

import (
	"believer/movies/db"
	"believer/movies/types"
	"believer/movies/utils"
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

	return c.Render("index", fiber.Map{
		"IsAdmin":  utils.IsAuthenticated(c),
		"Movies":   movies,
		"NextPage": page + 1,
	})
}

func HandleGetLogin(c *fiber.Ctx) error {
	return c.Render("login", fiber.Map{})
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
			Name:    "admin_secret",
			Value:   encoded,
			Expires: time.Now().AddDate(0, 0, 30),
		})
	}

	c.Set("HX-Redirect", "/")

	return c.SendStatus(200)
}
