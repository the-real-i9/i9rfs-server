package authcontrollers

import (
	"i9pkgs/i9auth"
	"log"
	"net/http"
	"os"
	"os/exec"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	connStream, err := websocket.Accept(w, r, wsOpts)
	if err != nil {
		return
	}

	defer connStream.CloseNow()

	var body struct {
		Step string
		Data any
	}

	for {
		r_err := wsjson.Read(r.Context(), connStream, &body)
		if r_err != nil {
			log.Println(r_err)
			return
		}

		switch body.Step {
		case "first", "one":
			email := body.Data.(string)

			var w_err error

			signupSessionJwt, app_err := i9auth.RequestNewAccount(email)
			if app_err != nil {
				w_err = wsjson.Write(r.Context(), connStream, map[string]any{"status": "f", "error": app_err.Error()})
			} else {
				w_err = wsjson.Write(r.Context(), connStream, map[string]any{"status": "s", "body": signupSessionJwt})
			}

			if w_err != nil {
				log.Println(w_err)
				return
			}

		case "second", "two":
			recvData := body.Data.(map[string]any)

			token := recvData["signup_session_jwt"].(string)
			code := recvData["code"].(int)

			var w_err error

			msg, app_err := i9auth.VerifyEmail(token, code)
			if app_err != nil {
				w_err = wsjson.Write(r.Context(), connStream, map[string]any{"status": "f", "error": app_err.Error()})
			} else {
				w_err = wsjson.Write(r.Context(), connStream, map[string]any{"status": "s", "body": msg})
			}

			if w_err != nil {
				log.Println(w_err)
				return
			}

		case "third", "three":
			recvData := body.Data.(map[string]any)

			token := recvData["signup_session_jwt"].(string)
			userInfo := recvData["user_info"].(map[string]any)

			var w_err error

			userData, jwtToken, app_err := i9auth.RegisterUser(token, userInfo, "")
			if app_err != nil {
				w_err = wsjson.Write(r.Context(), connStream, map[string]any{"status": "f", "error": app_err.Error()})
			} else {

				go createUserAccountDirectory(userData["username"].(string))

				respData := map[string]any{
					"msg":      "Signup Success!",
					"user":     userData,
					"auth_jwt": jwtToken,
				}

				w_err = wsjson.Write(r.Context(), connStream, map[string]any{"status": "s", "body": respData})
			}

			if w_err != nil {
				log.Println(w_err)
				return
			}
		}
	}
}

func createUserAccountDirectory(userAcc string) {
	fsHome := "i9FSHome"

	if hdir, app_err := os.UserCacheDir(); app_err == nil {
		fsHome = hdir + "/i9FSHome"
	}

	exec.Command("mkdir", "-p", fsHome+"/"+userAcc).Run()
}
