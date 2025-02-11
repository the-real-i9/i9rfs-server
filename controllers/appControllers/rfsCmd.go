package appControllers

import (
	"context"
	"fmt"
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"i9rfs/server/services/rfsCmdService"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RFSCmd = websocket.New(func(c *websocket.Conn) {
	user := c.Locals("user").(appTypes.ClientUser)

	var w_err error

	for {
		var body struct {
			Command  string
			WorkPath string
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

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		switch body.Command {
		case "cd":
			resp, app_err = rfsCmdService.ChangeDirectory(ctx, body.WorkPath, body.CmdArgs)
		case "mkdir":
			resp, app_err = rfsCmdService.MakeDirectory(ctx, body.WorkPath, body.CmdArgs, user.Username)
		case "rmdir":
			resp, app_err = rfsCmdService.RemoveDirectory(ctx, body.WorkPath, body.CmdArgs, user.Username)
		case "rm":
			resp, app_err = rfsCmdService.Remove(ctx, body.WorkPath, body.CmdArgs, user.Username)
		case "mv":
			resp, app_err = rfsCmdService.Move(body.WorkPath, body.CmdArgs)
		case "upload", "up":
			resp, app_err = rfsCmdService.UploadFile(body.WorkPath, body.CmdArgs)
		case "download", "down":
			resp, app_err = rfsCmdService.DownloadFile(body.WorkPath, body.CmdArgs)
		default:
			resp, app_err = nil, fmt.Errorf("unknown command: \"%s\"", body.Command)
		}

		if app_err != nil {
			w_err = c.WriteJSON(helpers.WSErrResp(app_err))
			continue
		}

		w_err = c.WriteJSON(appTypes.WSResp{
			StatusCode: fiber.StatusOK,
			Body:       resp,
		})
	}
})
