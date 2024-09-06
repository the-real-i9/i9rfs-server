package appRoutes

import (
	"i9rfs/server/src/controllers/appControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/session_user", appControllers.GetSessionUser)

	router.Get("/rfs", appControllers.RFSCmd)
}
