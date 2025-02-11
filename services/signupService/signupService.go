package signupService

import (
	"context"
	"fmt"
	"i9rfs/appTypes"
	user "i9rfs/models/userModel"
	"i9rfs/services/mailService"
	"i9rfs/services/securityServices"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RequestNewAccount(ctx context.Context, email string) (map[string]any, appTypes.SignupSession, error) {
	var session appTypes.SignupSession

	accExists, err := user.Exists(ctx, email)
	if err != nil {
		return nil, session, err
	}

	if accExists {
		return nil, session, fiber.NewError(400, fmt.Sprintf("signup error: an account with '%s' already exists", email))
	}

	verfCode, expires := securityServices.GetTokenAndExpiration()

	go mailService.SendMail(email, "Email Verification", fmt.Sprintf("Your email verification code is: <b>%d</b>", verfCode))

	session = appTypes.SignupSession{
		Step: "verify email",
		Data: appTypes.SignupSessionData{Email: email, VerificationCode: verfCode, VerificationCodeExpires: expires},
	}

	respData := map[string]any{
		"msg": "A 6-digit verification code has been sent to " + email,
	}

	return respData, session, nil
}

func VerifyEmail(ctx context.Context, sessionData appTypes.SignupSessionData, inputVerfCode int) (map[string]any, appTypes.SignupSession, error) {
	var updatedSession appTypes.SignupSession

	if sessionData.VerificationCode != inputVerfCode {
		return nil, updatedSession, fiber.NewError(fiber.StatusBadRequest, "email verification error: incorrect verification code")
	}

	if sessionData.VerificationCodeExpires.Before(time.Now()) {
		return nil, updatedSession, fiber.NewError(fiber.StatusBadRequest, "email verification error: verification code expired")
	}

	go mailService.SendMail(sessionData.Email, "Email Verification Success", fmt.Sprintf("Your email %s has been verified!", sessionData.Email))

	updatedSession = appTypes.SignupSession{
		Step: "register user",
		Data: appTypes.SignupSessionData{Email: sessionData.Email},
	}

	respData := map[string]any{
		"msg": fmt.Sprintf("Your email '%s' has been verified!", sessionData.Email),
	}

	return respData, updatedSession, nil
}

func RegisterUser(ctx context.Context, sessionData appTypes.SignupSessionData, username, password string) (any, string, error) {
	userExists, err := user.Exists(ctx, username)
	if err != nil {
		return nil, "", err
	}

	if userExists {
		return nil, "", fiber.NewError(fiber.StatusBadRequest, "signup error: username", username, "is unavailable")
	}

	hashedPassword, err := securityServices.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	newUser, err := user.New(ctx, sessionData.Email, username, hashedPassword)
	if err != nil {
		return nil, "", err
	}

	authJwt, err := securityServices.JwtSign(appTypes.ClientUser{
		Username: username,
	}, os.Getenv("AUTH_JWT_SECRET"), time.Now().UTC().Add(10*24*time.Hour)) // 1 year

	if err != nil {
		return nil, "", err
	}

	respData := map[string]any{
		"msg":  "Signup success!",
		"user": newUser,
	}

	return respData, authJwt, nil
}
