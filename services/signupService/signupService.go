package signupService

import (
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/appTypes"
	"i9rfs/server/models/appModel"
	user "i9rfs/server/models/userModel"
	"i9rfs/server/services/appServices"
	"i9rfs/server/services/mailService"
	"i9rfs/server/services/securityServices"
	"log"
	"os"
	"time"
)

func RequestNewAccount(email string) (string, error) {
	accExists, err := appModel.AccountExists(email)
	if err != nil {
		return "", err
	}

	if accExists {
		return "", fmt.Errorf("signup error: an account with '%s' already exists", email)
	}

	verfCode, expires := securityServices.GetTokenAndExpiration()

	sessionData := appTypes.SignupSessionData{
		Step:     "verify email",
		Email:    email,
		VerfCode: verfCode,
	}

	sessionId, err := appServices.NewSession("ongoing_signup", sessionData)
	if err != nil {
		return "", err
	}

	go mailService.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

	signupSessionJwt := securityServices.JwtSign(sessionId, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), expires)

	return signupSessionJwt, nil
}

func VerifyEmail(sessionId string, verfCode, inputVerfCode int, email string) (string, error) {
	if verfCode != inputVerfCode {
		return "", fmt.Errorf("email verification error: incorrect verification code")
	}

	sessionData := appTypes.SignupSessionData{
		Step:     "register user",
		Email:    email,
		VerfCode: 0,
	}

	err := appServices.UpdateSession("ongoing_signup", sessionId, sessionData)
	if err != nil {
		return "", err
	}

	go mailService.SendMail(email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", email))

	signupSessionJwt := securityServices.JwtSign(sessionId, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), time.Now().UTC().Add(1*time.Hour))

	return signupSessionJwt, nil
}

func RegisterUser(sessionId, email, username, password string) (any, error) {
	accExists, err := appModel.AccountExists(username)
	if err != nil {
		return nil, err
	}

	if accExists {
		return nil, fmt.Errorf("username error: username '%s' is unavailable", username)
	}

	hashedPassword, err := securityServices.HashPassword(password)
	if err != nil {
		log.Println(fmt.Errorf("authServices.go: RegisterUser: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	user, err := user.New(email, username, hashedPassword)
	if err != nil {
		return nil, err
	}

	userData := &appTypes.ClientUser{
		Id:       user.Id,
		Username: user.Username,
	}

	authJwt := securityServices.JwtSign(userData, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	appServices.EndSession("ongoing_signup", sessionId)

	respData := map[string]any{
		"msg":     "Signup success!",
		"user":    userData,
		"authJwt": authJwt,
	}

	return respData, nil
}
