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
	opts := &websocket.AcceptOptions{OriginPatterns: []string{"localhost"}}
	connStream, err := websocket.Accept(w, r, opts)
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
			signupSessionJwt, err := i9auth.RequestNewAccount(email)
			if err != nil {
				w_err = wsjson.Write(r.Context(), connStream, err.Error())
			} else {
				w_err = wsjson.Write(r.Context(), connStream, signupSessionJwt)
			}

			if w_err != nil {
				log.Println(w_err)
				return
			}

		case "second", "two":
			token := r.Header.Get("Authorization")
			code := body.Data.(int)

			var w_err error
			msg, err := i9auth.VerifyEmail(token, code)
			if err != nil {
				w_err = wsjson.Write(r.Context(), connStream, err.Error())
			} else {
				w_err = wsjson.Write(r.Context(), connStream, msg)
			}

			if w_err != nil {
				log.Println(w_err)
				return
			}

		case "third", "three":
			token := r.Header.Get("Authorization")
			userInfo := body.Data.(map[string]any)

			var w_err error
			userData, jwtToken, err := i9auth.RegisterUser(token, userInfo, "")
			if err != nil {
				w_err = wsjson.Write(r.Context(), connStream, err.Error())
			} else {

				go createUserAccountDirectory(userData["username"].(string))

				respData := map[string]any{
					"msg":      "Signup Success!",
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
}

func createUserAccountDirectory(userAcc string) {
	fsHome := "i9FSHome"

	if hdir, err := os.UserHomeDir(); err == nil {
		fsHome = hdir + "/i9FSHome"
	}

	exec.Command("mkdir", "-p", fsHome+"/"+userAcc).Run()
}
