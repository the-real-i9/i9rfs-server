package appRoutes

import (
	"i9rfs/server/appTypes"
	"i9rfs/server/controllers/appControllers"
	"i9rfs/server/services/securityServices"
	"os"

	"github.com/gofiber/fiber/v2"
)

func authenticateUser(c *fiber.Ctx) error {
	sessionToken := c.Get("Authorization")

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
	}

	c.Locals("user", clientUser)

	return c.Next()
}

func Init(router fiber.Router) {
	router.Get("/rfs", authenticateUser, appControllers.RFSCmd)
}
