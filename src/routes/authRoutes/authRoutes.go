package authRoutes

import (
	"i9rfs/src/controllers/auth/signinControllers"
	"i9rfs/src/controllers/auth/signupControllers"
	"i9rfs/src/middlewares/authMiddlewares"

	"github.com/gofiber/fiber/v2"
)

func Route(router fiber.Router) {
	router.Post("/signup/request_new_account", signupControllers.RequestNewAccount)
	router.Post("/signup/verify_email", authMiddlewares.SignupSession, signupControllers.VerifyEmail)
	router.Post("/signup/register_user", authMiddlewares.SignupSession, signupControllers.RegisterUser)

	router.Post("/signin", signinControllers.Signin)
}
