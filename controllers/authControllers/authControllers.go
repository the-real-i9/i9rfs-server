package authControllers

import (
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"i9rfs/server/services/authServices"
	"i9rfs/server/services/rfsCmdService"
	"log"
	"os"
	"os/exec"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber"
)

var Signup = websocket.New(func(c *websocket.Conn) {
	var w_err error

	for {
		var body signupBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {

			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		switch body.Step {
		case "one":
			resp := requestNewAccount(body.Data)

			w_err = c.WriteJSON(resp)

		case "two":
			sessionData, err := helpers.JwtVerify[appTypes.SignupSessionData](body.SessionToken, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))

			if err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			if sessionData.State != "verify email" {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			resp := verifyEmail(sessionData, body.Data)

			w_err = c.WriteJSON(resp)

		case "three":
			sessionData, err := helpers.JwtVerify[appTypes.SignupSessionData](body.SessionToken, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))

			if err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			if sessionData.State != "register user" {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			resp := registerUser(sessionData, body.Data)

			w_err = c.WriteJSON(resp)
		}
	}
})

func requestNewAccount(data map[string]any) any {

	var body requestNewAccountBody

	helpers.MapToStruct(data, &body)

	if val_err := body.Validate(); val_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err)
	}

	signupSessionJwt, app_err := authServices.RequestNewAccount(body.Email)

	if app_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err)
	}

	return appTypes.WSResp{
		StatusCode: fiber.StatusOK,
		Body:       signupSessionJwt,
	}
}

func verifyEmail(sessionData *appTypes.SignupSessionData, data map[string]any) any {
	var body verifyEmailBody

	helpers.MapToStruct(data, &body)

	if val_err := body.Validate(); val_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err)
	}

	signupSessionJwt, app_err := authServices.VerifyEmail(sessionData.SessionId, body.Code, sessionData.Email)

	if app_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err)
	}

	return appTypes.WSResp{
		StatusCode: fiber.StatusOK,
		Body:       signupSessionJwt,
	}
}

func registerUser(sessionData *appTypes.SignupSessionData, data map[string]any) any {
	var body registerUserBody

	helpers.MapToStruct(data, &body)

	if val_err := body.Validate(); val_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err)
	}

	userData, authJwt, app_err := authServices.RegisterUser(sessionData.SessionId, sessionData.Email, body.Username, body.Password)

	if app_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err)
	}

	go createUserAccountDirectory(userData.Username)

	return appTypes.WSResp{
		StatusCode: fiber.StatusOK,
		Body: map[string]any{
			"msg":     "Signup success!",
			"user":    userData,
			"authJwt": authJwt,
		},
	}
}

var Login = websocket.New(func(c *websocket.Conn) {

	var w_err error

	for {
		var body signInBody

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		if val_err := body.Validate(); val_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err))
			continue
		}

		userData, authJwt, app_err := authServices.Login(body.EmailOrUsername, body.Password)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		go createUserAccountDirectory(userData.Username)

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":     "Login success!",
				"user":    userData,
				"authJwt": authJwt,
			},
		})
	}
})

func createUserAccountDirectory(userAcc string) {
	fsHome := rfsCmdService.GetHome()

	exec.Command("mkdir", "-p", fsHome+"/"+userAcc).Run()
}
