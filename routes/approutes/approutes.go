package approutes

import (
	"i9rfs/server/controllers/appcontrollers"
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

func Init(router fiber.Router) {
	router.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("AUTH_JWT_SECRET"))},
		ContextKey: "auth",
	}))

	router.Get("/session_user", appcontrollers.GetSessionUser)

	router.Get("/rfs", appcontrollers.RFSCmd)
}
