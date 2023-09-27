package utils

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

func IsAuthenticated(c *fiber.Ctx) bool {
	adminSecret := os.Getenv("ADMIN_SECRET")
	cookieAdminSecret := c.Cookies("admin_secret")

	return cookieAdminSecret == adminSecret
}
