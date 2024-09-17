package appControllers

import (
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetSessionUser = websocket.New(func(c *websocket.Conn) {
	sessionToken := c.Headers("Authorization")

	user, err := helpers.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))

	if err != nil {
		w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
		if w_err != nil {
			log.Println(w_err)
		}
		return
	}

	var body struct{}

	for {
		w_err := c.WriteJSON(appTypes.WSResp{StatusCode: fiber.StatusOK, Body: map[string]any{"user": user}})
		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}
	}
})
