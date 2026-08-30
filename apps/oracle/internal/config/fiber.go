package config

import "github.com/gofiber/fiber/v2"

func NewFiber() *fiber.App {
	app := fiber.New()
	return app
}
