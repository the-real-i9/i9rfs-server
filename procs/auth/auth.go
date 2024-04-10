package authprocs

import (
	"encoding/json"
	"i9pkgs/i9auth"
)

type Auth struct{}

func (au Auth) Login(args struct {
	EmailOrUsername string
	Password        string
}, reply *string) error {

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

	return nil
}

func (au Auth) GetSessionUser(token string, reply *string) error {
	userData, err := i9auth.GetSessionUser(token)
	if err != nil {
		return err
	}

	respData, _ := json.Marshal(userData)

	*reply = string(respData)

	return nil
}
