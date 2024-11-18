package authRoutes

import (
	"i9rfs/server/controllers/auth/loginControllers"
	"i9rfs/server/controllers/auth/signupControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup", signupControllers.Signup)

	router.Get("/login", loginControllers.Login)
}
