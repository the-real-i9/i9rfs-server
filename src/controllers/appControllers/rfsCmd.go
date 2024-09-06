package appControllers

import (
	"i9rfs/server/src/appTypes"
	"i9rfs/server/src/helpers"
	"i9rfs/server/src/services/rfsCmdService"
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RFSCmd = websocket.New(func(c *websocket.Conn) {
	sessionToken := c.Headers("Authorization")

	_, err := helpers.JwtVerify[appTypes.ClientUser](sessionToken, os.Getenv("AUTH_JWT_SECRET"))

	if err != nil {
		w_err := c.WriteJSON(helpers.ErrResp(fiber.StatusUnauthorized, err))
		if w_err != nil {
			log.Println(w_err)
		}
		return
	}

	var w_err error

	for {
		var body struct {
			WorkPath string
			Command  string
			CmdArgs  []string
		}

		if w_err != nil {
			log.Println(w_err)
			break
		}

		r_err := c.ReadJSON(&body)
		if r_err != nil {
			log.Println(r_err)
			break
		}

		var (
			resp    any
			app_err error
		)

		switch body.Command {
		case "pex":
			resp, app_err = rfsCmdService.PathExists(body.WorkPath)
		case "upload", "up":
			resp, app_err = rfsCmdService.UploadFile(body.WorkPath, body.CmdArgs)
		case "download", "down":
			resp, app_err = rfsCmdService.DownloadFile(body.WorkPath, body.CmdArgs)
		default:
			resp, app_err = rfsCmdService.BashCommand(body.WorkPath, body.Command, body.CmdArgs)
		}

		if app_err != nil {
			w_err = c.WriteJSON(helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       resp,
		})
	}
})
