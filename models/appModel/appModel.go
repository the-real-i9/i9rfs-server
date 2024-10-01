package appModel

import (
	"context"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func AccountExists(emailOrUsername string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := appGlobals.DB.Collection("user")

	count, err := coll.CountDocuments(ctx, bson.M{"$or": bson.A{bson.M{"email": emailOrUsername}, bson.M{"username": emailOrUsername}}})
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
	result, err := helpers.MultiOpQuery(func(ctx context.Context) (any, error) {
		coll := appGlobals.DB.Collection("ongoing_signup")

		_, err := coll.DeleteOne(ctx, bson.M{"email": email})
		if err != nil {
			return nil, err
		}

		res, err := coll.InsertOne(ctx, bson.M{"email": email, "verification_code": verfCode, "verified": false})
		if err != nil {
			return nil, err
		}

		return res.InsertedID, nil
	})

	if err != nil {
		log.Println(fmt.Errorf("appModel.go: NewSignupSession: %s", err))
		return "", appGlobals.ErrInternalServerError
	}

	sessionId := result.(bson.ObjectID).String()

	return sessionId, nil
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
