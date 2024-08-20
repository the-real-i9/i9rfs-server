package appcontrollers

import (
	"i9rfs/server/appTypes"
	user "i9rfs/server/models/userModel"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var GetSessionUser = websocket.New(func(c *websocket.Conn) {
	user := c.Locals("user").(user.User)

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
