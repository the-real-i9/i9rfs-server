package authcontrollers

import (
	"i9pkgs/i9auth"
	"log"
	"net/http"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func Login(w http.ResponseWriter, r *http.Request) {
	opts := &websocket.AcceptOptions{OriginPatterns: []string{"localhost"}}
	connStream, err := websocket.Accept(w, r, opts)
	if err != nil {
		return
	}

	defer connStream.CloseNow()

	var body struct {
		EmailOrUsername string
		Password        string
	}

	for {
		r_err := wsjson.Read(r.Context(), connStream, &body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		var w_err error
		userData, jwtToken, app_err := i9auth.Login(body.EmailOrUsername, body.Password, "")
		if app_err != nil {
			w_err = wsjson.Write(r.Context(), connStream, app_err.Error())
		} else {

			respData := map[string]any{
				"msg":      "You're logged in!",
				"user":     userData,
				"auth_jwt": jwtToken,
			}

			w_err = wsjson.Write(r.Context(), connStream, respData)
		}

		if w_err != nil {
			log.Println(w_err)
			return
		}
	}
}
