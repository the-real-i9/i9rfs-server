package authRoutes

import (
	"i9rfs/server/controllers/authControllers"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup", authControllers.Signup)

	router.Get("/signin", authControllers.Signin)
}
