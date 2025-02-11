package middlewares

import (
	"i9rfs/server/appGlobals"
	"i9rfs/server/appTypes"
	"i9rfs/server/services/securityServices"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func Auth(c *fiber.Ctx) error {
	sess, err := appGlobals.UserSessionStore.Get(c)
	if err != nil {
		log.Println("auth.go: Auth: UserSignupSession.Get:", err)
		return fiber.ErrInternalServerError
	}

	sessionToken, ok := sess.Get("authJwt").(string)
	if !ok {
		log.Println("auth.go: Auth: sess.Get: authJwt is missing")
		return fiber.ErrInternalServerError

	}

	clientUser, err := securityServices.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))
	if err != nil {
		return err
	}

	c.Locals("user", clientUser)

	return c.Next()
}
