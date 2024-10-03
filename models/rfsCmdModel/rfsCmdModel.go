package rfsCmdModel

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

func PathExists(path string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := appGlobals.DB.Collection("directory").FindOne(ctx, bson.M{"path": path}).Decode(&struct{}{})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return false, nil
		}

		log.Println(fmt.Errorf("rmsCmdModel.go: PathExists: %s", err))
		return false, appGlobals.ErrInternalServerError
	}

	return true, nil
}
