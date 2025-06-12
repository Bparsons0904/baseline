package routes

import (
	"server/config"

	"github.com/gofiber/fiber/v2"
)

func HealthRoutes(router fiber.Router, config config.Config) {
	router.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": config.GeneralVersion,
			"service": "app_api",
		})
	})
}
