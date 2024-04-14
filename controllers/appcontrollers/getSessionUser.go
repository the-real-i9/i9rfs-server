package appcontrollers

import (
	"context"
	"i9pkgs/i9auth"
	"log"
	"net/http"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func GetSessionUser(w http.ResponseWriter, r *http.Request) {
	connStream, err := websocket.Accept(w, r, wsOpts)
	if err != nil {
		return
	}

	defer connStream.CloseNow()

	token := r.Header.Get("Authorization")

	var w_err error

	userData, app_err := i9auth.GetSessionUser(token)
	if app_err != nil {
		w_err = wsjson.Write(context.Background(), connStream, map[string]any{"status": "f", "error": app_err.Error()})
	} else {
		w_err = wsjson.Write(context.Background(), connStream, map[string]any{"status": "s", "body": userData})
	}

	if w_err != nil {
		log.Println(w_err)
		return
	}
}
