package authroutes

import (
	"i9rfs/server/controllers/authcontrollers"
	"i9rfs/server/middlewares"
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Get("/signup/request_new_account", authcontrollers.RequestNewAccount)

	router.Use("/signup", jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("SIGNUP_SESSION_JWT_SECRET"))},
		ContextKey: "auth",
	}))
	router.Get("/signup/verify_email", middlewares.VerifyEmail, authcontrollers.VerifyEmail)
	router.Get("/signup/register_user", middlewares.RegisterUser, authcontrollers.RegisterUser)

	router.Get("/signin", authcontrollers.Signin)
}
