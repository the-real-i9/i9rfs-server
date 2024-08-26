package main

import (
	"i9rfs/server/initializers"
	"i9rfs/server/routes/appRoutes"
	"i9rfs/server/routes/authRoutes"

	"net/http"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func init() {
	initializers.InitApp()
}

func main() {
	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Use("/", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Route("/api/auth", authRoutes.Init)

	app.Route("/api/app", appRoutes.Init)

	http.ListenAndServe(":8000", nil)
}
