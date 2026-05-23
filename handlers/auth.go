package handlers

import (
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	repo db.AuthQuerier
}

func NewAuthHandler(repo db.AuthQuerier) *AuthHandler {
	return &AuthHandler{repo}
}

type authFormData struct {
	Password string `form:"password"`
	Username string `form:"username"`
}

// Display the login view
func (h *AuthHandler) GetLogin(c *fiber.Ctx) error {
	return utils.Render(c, views.Login())
}

// Login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var user db.UserAuth

	data := new(authFormData)

	// Parse the form data
	if err := c.BodyParser(data); err != nil {
		return err
	}

	// Get the password hash of the user from the database
	var err error
	user, err = h.repo.GetUserForLogin(data.Username)

	if err != nil {
		return setHXError(c, "Invalid username or password")
	}

	// Check if the password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(data.Password))

	if err != nil {
		return setHXError(c, "Invalid username or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  user.ID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("ADMIN_SECRET")))

	if err != nil {
		return setHXError(c, "Something went wrong")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  time.Now().AddDate(0, 0, 30),
		HTTPOnly: true,
		Secure:   c.Locals("AppEnv") != "development",
	})

	return c.Redirect("/", 303)
}

// Logout and remove the token from the cookie
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().AddDate(0, 0, -1),
		HTTPOnly: true,
	})

	c.Set("HX-Redirect", "/")

	return c.SendStatus(200)
}

// Route to create a new account.
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	data := new(authFormData)

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

	err = h.repo.CreateUser(data.Username, string(hash))

	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func setHXError(c *fiber.Ctx, message string) error {
	c.Set("HX-Retarget", "#error")
	c.Set("HX-Reswap", "innerHTML")

	return c.Status(fiber.StatusUnauthorized).SendString(message)
}
