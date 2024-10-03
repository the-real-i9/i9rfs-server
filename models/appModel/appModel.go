package appModel

import (
	"context"
	"errors"
	"fmt"
	"i9rfs/server/appGlobals"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func AccountExists(emailOrUsername string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := appGlobals.DB.Collection("user").FindOne(ctx, bson.M{"$or": bson.A{bson.M{"email": emailOrUsername}, bson.M{"username": emailOrUsername}}}).Decode(&struct{}{})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		log.Println(fmt.Errorf("appModel.go: AccountExists: %s", err))
		return false, appGlobals.ErrInternalServerError
	}

	return true, nil
}
