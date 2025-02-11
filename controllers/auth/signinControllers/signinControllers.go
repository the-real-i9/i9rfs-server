package signinControllers

import (
	"context"
	"i9rfs/appGlobals"
	"i9rfs/services/signinService"
	"log"

	"github.com/gofiber/fiber/v2"
)

func Signin(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var body signinBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return val_err
	}

	respData, authJwt, app_err := signinService.Signin(ctx, body.EmailOrUsername, body.Password)
	if app_err != nil {
		return app_err
	}

	userSess, err := appGlobals.UserSessionStore.Get(c)
	if err != nil {
		log.Println("signinControllers.go: Signin: UserSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	userSess.Set("authJwt", authJwt)

	if err := userSess.Save(); err != nil {
		log.Println("signinControllers.go: Signin: userSess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}
