package app

import "github.com/gofiber/fiber/v2"

func mapURLs(r *fiber.App) {
	r.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(`{"status":"ok"}`)
	})

	r.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("healthy")
	})

	r.Post("/acquisition", handleAcquisition)
}

func handleAcquisition(c *fiber.Ctx) error {
	return c.JSON("")
}
