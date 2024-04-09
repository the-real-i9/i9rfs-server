package authprocs

import (
	"encoding/json"
	"fmt"
	"i9Packages/i9auth"
	"os/exec"
)

type AuthSignup struct{}

func (aus AuthSignup) RequestNewAccount(email string, reply *string) error {
	signupSessionJwtToken, err := i9auth.RequestNewAccount(email)
	if err != nil {
		return err
	}

	*reply = signupSessionJwtToken

	return nil
}

func (aus AuthSignup) VerifyEmail(args struct {
	Token string
	Code  int
}, reply *string) error {

	msg, err := i9auth.VerifyEmail(args.Token, args.Code)
	if err != nil {
		return err
	}

	*reply = msg

	return nil
}

func (aus AuthSignup) RegisterUser(args struct {
	Token    string
	UserInfo map[string]any
}, reply *string) error {

	userData, jwtToken, err := i9auth.RegisterUser(args.Token, args.UserInfo, "")
	if err != nil {
		return err
	}

	go exec.Command("mkdir", "-p", fmt.Sprintf("i9FSHome/%s", userData["username"])).Run()

	respData, _ := json.Marshal(map[string]any{
		"msg":        "Signup Success!",
		"user":       userData,
		"auth_token": jwtToken,
	})

	*reply = string(respData)

	return nil
}
