package authServices

import (
	"errors"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"i9rfs/server/models/appModel"
	"i9rfs/server/models/authModel"
	user "i9rfs/server/models/userModel"
	"i9rfs/server/services/appServices"
	"log"
	"math/rand"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func RequestNewAccount(email string) (string, error) {
	accExists, err := appModel.AccountExists(email)
	if err != nil {
		return "", err
	}

	if accExists {
		return "", fmt.Errorf("signup error: an account with '%s' already exists", email)
	}

	verfCode, expires := rand.Intn(899999)+100000, time.Now().UTC().Add(1*time.Hour)

	sessionId, err := authModel.NewSignupSession(email, verfCode)
	if err != nil {
		return "", err
	}

	go appServices.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

	signupSessionJwt := helpers.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "verify email",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), expires)

	return signupSessionJwt, nil
}

func VerifyEmail(sessionId string, inputVerfCode int, email string) (string, error) {
	isSuccess, err := authModel.VerifyEmail(sessionId, inputVerfCode)
	if err != nil {
		return "", err
	}

	if !isSuccess {
		return "", fmt.Errorf("email verification error: incorrect verification code")
	}

	go appServices.SendMail(email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", email))

	signupSessionJwt := helpers.JwtSign(appTypes.SignupSessionData{
		SessionId: sessionId,
		Email:     email,
		Step:      "register user",
	}, os.Getenv("SIGNUP_SESSION_JWT_SECRET"), time.Now().UTC().Add(1*time.Hour))

	return signupSessionJwt, nil
}

func RegisterUser(sessionId, email, username, password string) (*appTypes.ClientUser, string, error) {
	accExists, err := appModel.AccountExists(username)
	if err != nil {
		return nil, "", err
	}

	if accExists {
		return nil, "", fmt.Errorf("username error: username '%s' is unavailable", username)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(fmt.Errorf("authServices.go: RegisterUser: %s", err))
		return nil, "", appGlobals.ErrInternalServerError
	}

	user, err := user.New(email, username, string(hashedPassword))
	if err != nil {
		return nil, "", err
	}

	clientUser := &appTypes.ClientUser{
		Id:       user.Id,
		Username: user.Username,
	}

	authJwt := helpers.JwtSign(clientUser, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour)) // 1 year

	authModel.EndSignupSession(sessionId)

	return clientUser, authJwt, nil
}

func Login(emailOrUsername, password string) (*appTypes.ClientUser, string, error) {
	user, err := user.FindOne(emailOrUsername)
	if err != nil {
		return nil, "", err
	}

	if user == nil {
		return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
	}

	cmp_err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if cmp_err != nil {
		if errors.Is(cmp_err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, "", fmt.Errorf("signin error: incorrect email/username or password")
		} else {
			log.Println(fmt.Errorf("authServices.go: Signin: %s", cmp_err))
			return nil, "", appGlobals.ErrInternalServerError
		}
	}

	clientUser := &appTypes.ClientUser{
		Id:       user.Id,
		Username: user.Username,
	}

	authJwt := helpers.JwtSign(clientUser, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(365*24*time.Hour))

	return clientUser, authJwt, nil
}
