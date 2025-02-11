package authRoutes

import (
	"i9rfs/server/controllers/auth/signinControllers"
	"i9rfs/server/controllers/auth/signupControllers"
	"i9rfs/server/middlewares/ssm"

	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup/request_new_account", signupControllers.RequestNewAccount)
	router.Get("/signup/verify_email", ssm.VerifyEmail, signupControllers.VerifyEmail)
	router.Get("/signup/register_user", ssm.RegisterUser, signupControllers.RegisterUser)

	router.Get("/signin", signinControllers.Signin)
}
