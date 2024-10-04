package appControllers

import (
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"i9rfs/server/services/rfsCmdService"
	"log"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RFSCmd = websocket.New(func(c *websocket.Conn) {
	user := c.Locals("user").(*appTypes.ClientUser)

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
		case "cd":
			resp, app_err = rfsCmdService.ChangeDirectory(body.WorkPath, body.CmdArgs)
		case "mkdir":
			resp, app_err = rfsCmdService.MakeDirectory(body.WorkPath, body.CmdArgs, user.Id)
		case "upload", "up":
			resp, app_err = rfsCmdService.UploadFile(body.WorkPath, body.CmdArgs)
		case "download", "down":
			resp, app_err = rfsCmdService.DownloadFile(body.WorkPath, body.CmdArgs)
		default:
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
