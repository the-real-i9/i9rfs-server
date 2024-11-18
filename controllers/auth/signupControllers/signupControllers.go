package signupControllers

import (
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"i9rfs/server/services/appServices"
	"i9rfs/server/services/securityServices"
	"log"
	"os"

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
			sessionId, err := securityServices.JwtVerify[string](body.SessionToken, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
			if err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			sessionData, err := appServices.RetrieveSession[appTypes.SignupSessionData]("ongoing_signup", *sessionId)
			if err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			if sessionData.Step != "verify email" {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			resp := verifyEmail(*sessionId, sessionData, body.Data)

			w_err = c.WriteJSON(resp)

		case "three":
			sessionId, err := securityServices.JwtVerify[string](body.SessionToken, os.Getenv("SIGNUP_SESSION_JWT_SECRET"))
			if err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			sessionData, err := appServices.RetrieveSession[appTypes.SignupSessionData]("ongoing_signup", *sessionId)
			if err != nil {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			if sessionData.Step != "register user" {
				w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
				break
			}

			resp := registerUser(*sessionId, sessionData, body.Data)

			w_err = c.WriteJSON(resp)
		}
	}
})
