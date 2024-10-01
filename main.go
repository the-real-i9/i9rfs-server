package main

import (
	"context"
	"i9rfs/server/appGlobals"
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
	defer func() {
		if err := appGlobals.DB.Client().Disconnect(context.TODO()); err != nil {
			log.Panic(err)
		}
	}()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

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
