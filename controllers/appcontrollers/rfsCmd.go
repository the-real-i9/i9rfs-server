package appcontrollers

import (
	"fmt"
	"i9rfs/server/services/rfscmdservice"
	"log"
	"net/http"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func RFSCmd(w http.ResponseWriter, r *http.Request) {
	connStream, err := websocket.Accept(w, r, wsOpts)
	if err != nil {
		return
	}

	defer connStream.CloseNow()

	var body struct {
		WorkPath string
		Command  string
		CmdArgs  []string
	}

	for {
		r_err := wsjson.Read(r.Context(), connStream, &body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		var (
			resp    any
			app_err error
		)

		switch body.Command {
		case "cd":
			resp, app_err = rfscmdservice.PathExists(body.WorkPath)
		case "ls", "cat", "touch", "mkdir", "cp", "mv", "rm", "rmdir":
			resp, app_err = rfscmdservice.FileMgmtCommand(body.WorkPath, body.Command, body.CmdArgs)
		case "upload", "up":
			resp, app_err = rfscmdservice.UploadFile(body.WorkPath, body.CmdArgs)
		case "download", "down":
			resp, app_err = rfscmdservice.DownloadFile(body.WorkPath, body.CmdArgs)
		default:
			resp, app_err = "", fmt.Errorf("Command '%s' not found", body.Command)
		}

		var w_err error

		if app_err != nil {
			w_err = wsjson.Write(r.Context(), connStream, app_err.Error())
		} else {
			w_err = wsjson.Write(r.Context(), connStream, resp)
		}

		if w_err != nil {
			log.Println(w_err)
			return
		}
	}
}
