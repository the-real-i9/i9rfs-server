package authcontrollers

import (
	"fmt"
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"i9rfs/server/services/authservices"
	"log"
	"os"
	"os/exec"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber"
)

var RequestNewAccount = websocket.New(func(c *websocket.Conn) {

	var w_err error

	for {
		var body requestNewAccountBody

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

		signupSessionJwt, app_err := authservices.RequestNewAccount(body.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":           "A 6-digit verification code has been sent to " + body.Email,
				"session_token": signupSessionJwt,
			},
		})
	}
})

var VerifyEmail = websocket.New(func(c *websocket.Conn) {
	sessionData := c.Locals("signupSessionData").(appTypes.SignupSessionData)

	var w_err error

	for {
		var body verifyEmailBody

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

		signupSessionJwt, app_err := authservices.VerifyEmail(sessionData.SessionId, body.Code, sessionData.Email)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":           fmt.Sprintf("Your email '%s' has been verified!", sessionData.Email),
				"session_token": signupSessionJwt,
			},
		})
	}
})

var RegisterUser = websocket.New(func(c *websocket.Conn) {
	sessionData := c.Locals("signupSessionData").(appTypes.SignupSessionData)

	var w_err error

	for {
		var body registerUserBody

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

		userData, authJwt, app_err := authservices.RegisterUser(sessionData.SessionId, sessionData.Email, body.Username, body.Password)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		go createUserAccountDirectory(userData.Username)

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":     "Signup success!",
				"user":    userData,
				"authJwt": authJwt,
			},
		})
	}
})

var Signin = websocket.New(func(c *websocket.Conn) {

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

		userData, authJwt, app_err := authservices.Signin(body.EmailOrUsername, body.Password)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		go createUserAccountDirectory(userData.Username)

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body: map[string]any{
				"msg":     "Signin success!",
				"user":    userData,
				"authJwt": authJwt,
			},
		})
	}
})

func createUserAccountDirectory(userAcc string) {
	fsHome := "i9FSHome"

	if hdir, app_err := os.UserCacheDir(); app_err == nil {
		fsHome = hdir + "/i9FSHome"
	}

	exec.Command("mkdir", "-p", fsHome+"/"+userAcc).Run()
}
