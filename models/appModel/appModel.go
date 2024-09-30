package appModel

import (
	"context"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func AccountExists(emailOrUsername string) (bool, error) {
	count, err := appGlobals.DB.Collection("user").CountDocuments(context.TODO(), bson.M{"$or": bson.A{bson.M{"email": emailOrUsername}, bson.M{"username": emailOrUsername}}})
	if err != nil {
		log.Println(fmt.Errorf("appModel.go: NewSignupSession: %s", err))
		return false, appGlobals.ErrInternalServerError
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}

func NewSignupSession(email string, verfCode int) (string, error) {
	sessionId, err := helpers.QueryRowField[string]("SELECT session_id FROM new_signup_session($1, $2)", email, verfCode)

	if err != nil {
		log.Println(fmt.Errorf("appModel.go: NewSignupSession: %s", err))
		return "", appGlobals.ErrInternalServerError
	}

	return *sessionId, nil
}

func VerifyEmail(sessionId string, verfCode int) (bool, error) {
	isSuccess, err := helpers.QueryRowField[bool]("SELECT is_success FROM verify_email($1, $2)", sessionId, verfCode)

	if err != nil {
		log.Println(fmt.Errorf("appModel.go: VerifyEmail: %s", err))
		return false, appGlobals.ErrInternalServerError
	}

	return *isSuccess, nil
}

func EndSignupSession(sessionId string) {
	go helpers.QueryRowField[bool]("SELECT end_signup_session ($1)", sessionId)
}
