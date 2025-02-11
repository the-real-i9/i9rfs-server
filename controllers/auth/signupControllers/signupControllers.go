package signupControllers

import (
	"context"
	"encoding/json"
	"i9rfs/server/appGlobals"
	"i9rfs/server/appTypes"
	"i9rfs/server/services/signupService"
	"log"

	"github.com/gofiber/fiber/v2"
)

func RequestNewAccount(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var body requestNewAccountBody

	if b_err := c.BodyParser(&body); b_err != nil {
		return b_err
	}

	if val_err := body.Validate(); val_err != nil {
		return fiber.NewError(400, val_err.Error())
	}

	respData, session, app_err := signupService.RequestNewAccount(ctx, body.Email)
	if app_err != nil {
		return app_err
	}

	signupSess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: RequestNewAccount: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	sd, err := json.Marshal(session)
	if err != nil {
		log.Println("signupControllers.go: RequestNewAccount: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	signupSess.Set("signup_session", sd)

	if err := signupSess.Save(); err != nil {
		log.Println("signupControllers.go: RequestNewAccount: signupSess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func VerifyEmail(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionData := c.Locals("signup_session_data").(appTypes.SignupSessionData)

	var body verifyEmailBody

	if b_err := c.BodyParser(&body); b_err != nil {
		return b_err
	}

	if val_err := body.Validate(); val_err != nil {
		return fiber.NewError(400, val_err.Error())
	}

	respData, updatedSession, app_err := signupService.VerifyEmail(ctx, sessionData, body.Code)
	if app_err != nil {
		return app_err
	}

	signupSess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: VerifyEmail: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	usd, err := json.Marshal(updatedSession)
	if err != nil {
		log.Println("signupControllers.go: VerifyEmail: json.Marshal:", err)
		return fiber.ErrInternalServerError
	}

	signupSess.Set("signup_session", usd)

	if err := signupSess.Save(); err != nil {
		log.Println("signupControllers.go: VerifyEmail: signupSess.Save:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}

func RegisterUser(c *fiber.Ctx) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionData := c.Locals("signup_session_data").(appTypes.SignupSessionData)

	var body registerUserBody

	body_err := c.BodyParser(&body)
	if body_err != nil {
		return body_err
	}

	if val_err := body.Validate(); val_err != nil {
		return fiber.NewError(400, val_err.Error())
	}

	respData, authJwt, app_err := signupService.RegisterUser(ctx, sessionData, body.Username, body.Password)
	if app_err != nil {
		return app_err
	}

	userSess, err := appGlobals.UserSessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: RegisterUser: UserSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	userSess.Set("authJwt", authJwt)

	if err := userSess.Save(); err != nil {
		log.Println("signupControllers.go: RegisterUser: userSess.Save:", err)
		return fiber.ErrInternalServerError
	}

	signupSess, err := appGlobals.SignupSessionStore.Get(c)
	if err != nil {
		log.Println("signupControllers.go: RegisterUser: SignupSessionStore.Get:", err)
		return fiber.ErrInternalServerError
	}

	if err := signupSess.Destroy(); err != nil {
		log.Println("signupControllers.go: RegisterUser: signupSess.Destroy:", err)
		return fiber.ErrInternalServerError
	}

	return c.JSON(respData)
}
