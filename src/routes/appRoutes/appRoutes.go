package appRoutes

import (
	"i9rfs/src/controllers/appControllers"
	"i9rfs/src/middlewares/authMiddlewares"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Use(authMiddlewares.UserAuth)

	router.Get("/signout")

	router.Use("/rfs", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	router.Get("/rfs", appControllers.RFSAction)
}
