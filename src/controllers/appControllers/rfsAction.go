package appControllers

import (
	"context"
	"fmt"
	"i9rfs/src/appTypes"
	"i9rfs/src/helpers"
	"i9rfs/src/services/rfsActionService"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

func Signout(c *fiber.Ctx) error {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	c.ClearCookie()

	return c.JSON(fmt.Sprintf("Bye, %s! See you again!", clientUser.Username))
}

var RFSAction = websocket.New(func(c *websocket.Conn) {
	clientUser := c.Locals("user").(appTypes.ClientUser)

	var w_err error

	for {
		var body rfsActionBody

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
			c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
			continue
		}

		var (
			resp    any
			app_err error
		)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		switch body.Action {
		case "list directory contents", "ls":
			var data lsAction

			helpers.ToStruct(body.Data, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
				continue
			}

			resp, app_err = rfsActionService.Ls(ctx, clientUser.Username, data.DirectoryId)
		case "create new directory", "mkdir":
			var data mkdirAction

			helpers.ToStruct(body.Data, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
				continue
			}

			resp, app_err = rfsActionService.Mkdir(ctx, clientUser.Username, data.ParentDirectoryId, data.DirectoryName)
		case "delete", "del":
			var data delAction

			helpers.ToStruct(body.Data, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
				continue
			}

			resp, app_err = rfsActionService.Del(ctx, clientUser.Username, data.ParentDirectoryId, data.ObjectIds)
		case "trash":
			var data trashAction

			helpers.ToStruct(body.Data, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
				continue
			}

			resp, app_err = rfsActionService.Trash(ctx, clientUser.Username, data.ParentDirectoryId, data.ObjectIds)
		case "restore":
			var data restoreAction

			helpers.ToStruct(body.Data, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
				continue
			}

			resp, app_err = rfsActionService.Restore(ctx, clientUser.Username, data.ObjectIds)
		case "show trash", "view trash":
			resp, app_err = rfsActionService.ShowTrash(ctx, clientUser.Username)
		case "rename":
			var data renameAction

			helpers.ToStruct(body.Data, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
				continue
			}

			resp, app_err = rfsActionService.Rename(ctx, clientUser.Username, data.ParentDirectoryId, data.ObjectId, data.NewName)
		case "move":
			var data moveAction

			helpers.ToStruct(body.Data, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
				continue
			}

			resp, app_err = rfsActionService.Move(ctx, clientUser.Username, data.FromParentDirectoryId, data.ToParentDirectoryId, data.ObjectIds)
		case "copy":
			var data copyAction

			helpers.ToStruct(body.Data, &data)

			if val_err := data.Validate(); val_err != nil {
				c.WriteJSON(helpers.WSErrReply(val_err, body.Action))
				continue
			}

			resp, app_err = rfsActionService.Copy(ctx, clientUser.Username, data.FromParentDirectoryId, data.ToParentDirectoryId, data.ObjectIds)
		case "upload", "up":
			resp, app_err = nil, nil
		default:
			resp, app_err = nil, fmt.Errorf("unknown action: \"%s\"", body.Action)
		}

		if app_err != nil {
			w_err = c.WriteJSON(helpers.WSErrReply(app_err, body.Action))
			continue
		}

		w_err = c.WriteJSON(helpers.WSReply(resp, body.Action))
	}
})
