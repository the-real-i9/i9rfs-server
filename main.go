package main

import (
	"i9rfs/server/initializers"
	"i9rfs/server/routes/appRoutes"
	"i9rfs/server/routes/authRoutes"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func init() {
	if err := initializers.InitApp(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	cleanup, err := initializers.InitDBClient()
	if err != nil {
		log.Fatalln(err)
	}

	defer cleanup()

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Route("/api/auth", authRoutes.Init)

	app.Route("/api/app", appRoutes.Init)

	app.Listen(":8000")
}
