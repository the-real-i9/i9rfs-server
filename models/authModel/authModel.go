package authModel

import (
	"context"
	"errors"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func NewSignupSession(email string, verfCode int) (string, error) {
	db := appGlobals.DB

	result, err := helpers.MultiOpQuery(db.Client(), func(ctx context.Context) (any, error) {
		coll := db.Collection("ongoing_signup")

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

	sessionId := result.(bson.ObjectID).Hex()

	return sessionId, nil
}

func VerifyEmail(sessionId string, verfCode int) (bool, error) {
	sessionIdOid, err := bson.ObjectIDFromHex(sessionId)
	if err != nil {
		log.Println(fmt.Errorf("appModel.go: VerifyEmail: %s", err))
		return false, appGlobals.ErrInternalServerError
	}

	db := appGlobals.DB

	result, err := helpers.MultiOpQuery(db.Client(), func(ctx context.Context) (any, error) {
		coll := db.Collection("ongoing_signup")

		// get verification code from coll
		res := coll.FindOneAndUpdate(ctx, bson.M{"_id": sessionIdOid, "verification_code": verfCode}, bson.M{"$set": bson.M{"verified": true}})
		if err := res.Decode(&struct{}{}); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return false, nil
			}

			return false, err
		}

		return true, nil
	})

	if err != nil {
		log.Println(fmt.Errorf("appModel.go: VerifyEmail: %s", err))
		return false, appGlobals.ErrInternalServerError
	}

	return result.(bool), nil
}

func EndSignupSession(sessionId string) {
	go func() {
		sessionIdOid, _ := bson.ObjectIDFromHex(sessionId)
		go appGlobals.DB.Collection("ongoing_signup").DeleteOne(context.TODO(), bson.M{"_id": sessionIdOid})
	}()
}
