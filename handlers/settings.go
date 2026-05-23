package handlers

import (
	"believer/movies/components/checkbox"
	"believer/movies/db"
	"believer/movies/utils"
	"believer/movies/views"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type SettingsHandler struct {
	repo db.SettingsQuerier
}

func NewSettingsHandler(repo db.SettingsQuerier) *SettingsHandler {
	return &SettingsHandler{repo}
}

func (h *SettingsHandler) Settings(c *fiber.Ctx) error {
	isAuthenticated := utils.IsAuthenticated(c)

	if !isAuthenticated {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userId := c.Locals("UserId").(string)
	storedProviders, err := h.repo.GetWatchProviders(userId)

	if err != nil {
		return err
	}

	return utils.Render(c, views.Settings(views.SettingsProps{
		Providers: watchProviders(storedProviders),
	}))
}

func (h *SettingsHandler) SettingsWatchProviders(c *fiber.Ctx) error {
	isAuthenticated := utils.IsAuthenticated(c)

	if !isAuthenticated {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	userId := c.Locals("UserId").(string)
	formData := new(struct {
		Providers []string `form:"providers"`
	})

	if err := c.BodyParser(formData); err != nil {
		return err
	}

	selectedProviders := strings.Join(formData.Providers, ",")

	err := h.repo.UpdateWatchProviders(userId, selectedProviders)

	if err != nil {
		return err
	}

	return utils.Render(c, views.Settings(views.SettingsProps{
		Providers: watchProviders(selectedProviders),
	}))
}

func watchProviders(selectedProviders string) []checkbox.Props {
	var providers []checkbox.Props

	companies := []string{
		"Amazon Video",
		"Apple TV",
		"Disney Plus",
		"HBO Max",
		"Netflix",
		"SF Anytime",
		"SVT",
		"TV4 Play",
		"Viaplay",
	}

	for _, c := range companies {
		providers = append(providers, checkbox.Props{
			Checked: strings.Contains(selectedProviders, c),
			ID:      strings.ToLower(utils.Slugify(c)),
			Label:   c,
			Name:    "providers",
			Value:   c,
		})
	}

	return providers
}
