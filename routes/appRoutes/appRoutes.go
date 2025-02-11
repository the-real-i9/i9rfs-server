package appRoutes

import (
	"i9rfs/controllers/appControllers"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	router.Get("/rfs", appControllers.RFSCmd)
}
