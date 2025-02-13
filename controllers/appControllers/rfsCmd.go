package appControllers

import (
	"context"
	"fmt"
	"i9rfs/appTypes"
	"i9rfs/helpers"
	"i9rfs/services/rfsCmdService"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var RFSCmd = websocket.New(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	var w_err error

	for {
		var body rfsCmdBody

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
			c.WriteJSON(helpers.WSErrResp(val_err))
			continue
		}

		var (
			resp    any
			app_err error
		)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		switch body.Command {
		case "list directory contents", "ls":
			var data lsCmd

			helpers.MapToStruct(body.CmdData, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrResp(val_err))
				continue
			}

			resp, app_err = rfsCmdService.Ls(ctx, clientUser.Username, data.DirectoryId)
		case "create new directory", "mkdir":
			var data mkdirCmd

			helpers.MapToStruct(body.CmdData, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrResp(val_err))
				continue
			}

			resp, app_err = rfsCmdService.Mkdir(ctx, clientUser.Username, data.ParentDirectoryId, data.DirectoryName)
		case "delete", "del":
			var data delCmd

			helpers.MapToStruct(body.CmdData, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrResp(val_err))
				continue
			}

			resp, app_err = rfsCmdService.Del(ctx, clientUser.Username, data.ParentDirectoryId, data.ObjectIds)
		case "trash":
			var data trashCmd

			helpers.MapToStruct(body.CmdData, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrResp(val_err))
				continue
			}

			resp, app_err = rfsCmdService.Trash(ctx, clientUser.Username, data.ParentDirectoryId, data.ObjectIds)
		case "restore":
			var data restoreCmd

			helpers.MapToStruct(body.CmdData, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrResp(val_err))
				continue
			}

			resp, app_err = rfsCmdService.Restore(ctx, clientUser.Username, data.ObjectIds)
		case "show trash", "view trash":
			resp, app_err = rfsCmdService.ShowTrash(ctx, clientUser.Username)
		case "rename":
			var data renameCmd

			helpers.MapToStruct(body.CmdData, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrResp(val_err))
				continue
			}

			resp, app_err = rfsCmdService.Rename(ctx, clientUser.Username, data.ParentDirectoryId, data.ObjectId, data.NewName)
		case "move":
			var data moveCmd

			helpers.MapToStruct(body.CmdData, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrResp(val_err))
				continue
			}

			resp, app_err = rfsCmdService.Move(ctx, clientUser.Username, data.FromParentDirectoryId, data.ToParentDirectoryId, data.ObjectIds)
		case "copy":
			var data copyCmd

			helpers.MapToStruct(body.CmdData, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrResp(val_err))
				continue
			}

			resp, app_err = rfsCmdService.Copy(ctx, clientUser.Username, data.FromParentDirectoryId, data.ToParentDirectoryId, data.ObjectIds)
		case "upload", "up":
			resp, app_err = nil, nil
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
