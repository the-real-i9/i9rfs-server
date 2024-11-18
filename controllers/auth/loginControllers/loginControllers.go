package loginControllers

import (
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"i9rfs/server/services/loginService"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber"
)

var Login = websocket.New(func(c *websocket.Conn) {

	var w_err error

	for {
		var body loginInBody

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

		respData, app_err := loginService.Login(body.EmailOrUsername, body.Password)

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       respData,
		})
	}
})
