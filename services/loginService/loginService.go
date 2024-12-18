package loginService

import (
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/appTypes"
	user "i9rfs/server/models/userModel"
	"i9rfs/server/services/securityServices"
	"log"
	"os"
	"time"
)

func Login(emailOrUsername, password string) (any, error) {
	user, err := user.FindByEmailOrUsername(emailOrUsername)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("signin error: incorrect email/username or password")
	}

	matches, err := securityServices.HashAndPasswordMatches(user.Password, password)
	if err != nil {
		log.Println(fmt.Errorf("authServices.go: Login: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	if !matches {
		return nil, fmt.Errorf("signin error: incorrect email/username or password")
	}

	userData := &appTypes.ClientUser{
		Id:       user.Id,
		Username: user.Username,
	}

	authJwt := securityServices.JwtSign(userData, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour))

	respData := map[string]any{
		"msg":     "Login success!",
		"user":    userData,
		"authJwt": authJwt,
	}

	return respData, nil
}
