package authcontrollers

import "net/http"

func Login(w http.ResponseWriter, r *http.Request) {
	/*
		 userData, jwtToken, err := i9auth.Login(args.EmailOrUsername, args.Password, "")
			if err != nil {
				return err
			}

			respData, _ := json.Marshal(map[string]any{
				"msg":        "You're logged in!",
				"user":       userData,
				"auth_token": jwtToken,
			})

			*reply = string(respData)
	*/
}
