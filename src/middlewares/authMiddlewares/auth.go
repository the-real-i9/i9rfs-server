package authMiddlewares

import (
	"encoding/json"
	"i9rfs/src/appTypes"
	"i9rfs/src/services/securityServices"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func UserAuth(c *fiber.Ctx) error {
	usStr := c.Cookies("user")

	if usStr == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("authentication required")
	}

	var userSessionData map[string]string

	if err := json.Unmarshal([]byte(usStr), &userSessionData); err != nil {
		log.Println("auth.go: UserAuth: json.Unmarshal:", err)
		return fiber.ErrInternalServerError
	}

	sessionToken := userSessionData["authJwt"]

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("user", clientUser)

	return c.Next()
}
