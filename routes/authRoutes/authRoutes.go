package authRoutes

import (
	"i9rfs/server/controllers/authControllers"
	"i9rfs/server/middlewares"
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup/request_new_account", authControllers.RequestNewAccount)

	router.Use("/signup", jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("SIGNUP_SESSION_JWT_SECRET"))},
		ContextKey: "auth",
	}))
	router.Get("/signup/verify_email", middlewares.VerifyEmail, authControllers.VerifyEmail)
	router.Get("/signup/register_user", middlewares.RegisterUser, authControllers.RegisterUser)

	router.Get("/signin", authControllers.Signin)
}
