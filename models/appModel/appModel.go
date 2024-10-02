package appModel

import (
	"context"
	"fmt"
	"i9rfs/server/appGlobals"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func AccountExists(emailOrUsername string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := appGlobals.DB.Collection("user").CountDocuments(ctx, bson.M{"$or": bson.A{bson.M{"email": emailOrUsername}, bson.M{"username": emailOrUsername}}})
	if err != nil {
		log.Println(fmt.Errorf("appModel.go: AccountExists: %s", err))
		return false, appGlobals.ErrInternalServerError
	}

	if count == 0 {
		return false, nil
	}

	return true, nil
}
