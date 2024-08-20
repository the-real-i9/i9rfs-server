package middlewares

import (
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func VerifyEmail(c *fiber.Ctx) error {
	jwtData := c.Locals("auth").(*jwt.Token).Claims.(jwt.MapClaims)["data"].(map[string]any)

	var signupSessionData appTypes.SignupSessionData

	helpers.MapToStruct(jwtData, &signupSessionData)

	if signupSessionData.State != "verify email" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	c.Locals("signupSessionData", signupSessionData)

	return c.Next()
}

func RegisterUser(c *fiber.Ctx) error {
	jwtData := c.Locals("auth").(*jwt.Token).Claims.(jwt.MapClaims)["data"].(map[string]any)

	var signupSessionData appTypes.SignupSessionData

	helpers.MapToStruct(jwtData, &signupSessionData)

	if signupSessionData.State != "register user" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	c.Locals("signupSessionData", signupSessionData)

	return c.Next()
}

func GetSessionUser(c *fiber.Ctx) error {
	jwtData := c.Locals("auth").(*jwt.Token).Claims.(jwt.MapClaims)["data"].(map[string]any)

	var user appTypes.ClientUser

	helpers.MapToStruct(jwtData, &user)

	c.Locals("user", user)

	return c.Next()
}
