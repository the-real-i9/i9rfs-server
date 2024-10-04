package rfsCmdModel

import (
	"context"
	"errors"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

func NewDirectory(parentDirPath string, newDirTree []string, userId string) (bool, error) {
	userOid, _ := bson.ObjectIDFromHex(userId)

	db := appGlobals.DB

	result, err := helpers.MultiOpQuery(db.Client(), func(ctx context.Context) (any, error) {
		var parentDir struct {
			Oid        bson.ObjectID `bson:"_id"`
			Properties struct {
				Path string `bson:"path"`
			} `bson:"properties"`
		}

		err := db.Collection("directory").FindOne(ctx, bson.M{"properties.path": bson.M{"$eq": parentDirPath}}, options.FindOne().SetProjection(bson.M{"_id": 1, "properties.path": 1})).Decode(&parentDir)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return false, err
		}

		dirDate := time.Now().Format("2 January 2006 3:04:05 PM")
		for _, dirName := range newDirTree {
			dirName := strings.Trim(dirName, "\"")
			var dir struct {
				Oid        bson.ObjectID `bson:"_id"`
				Properties struct {
					Path string `bson:"path"`
				} `bson:"properties"`
			}

			err := db.Collection("directory").FindOne(ctx, bson.M{"properties.path": bson.M{"$eq": parentDir.Properties.Path + "/" + dirName}}, options.FindOne().SetProjection(bson.M{"_id": 1, "properties.path": 1})).Decode(&dir)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return false, err
			}

			if dir.Properties.Path != "" {
				parentDir.Oid = dir.Oid
				parentDir.Properties.Path = dir.Properties.Path
				continue
			}

			dirDoc := bson.M{
				"owner_user_id": userOid,
				"properties": bson.M{
					"name":          dirName,
					"path":          parentDir.Properties.Path + "/" + dirName,
					"date_modified": dirDate,
					"date_created":  dirDate,
				},
			}

			if parentDir.Properties.Path != "" {
				dirDoc["parent_directory_id"] = parentDir.Oid
			}

			res, err := db.Collection("directory").InsertOne(ctx, dirDoc)
			if err != nil {
				return nil, err
			}

			// this new directory will be the parent of the next
			parentDir.Oid = res.InsertedID.(bson.ObjectID)
			parentDir.Properties.Path = parentDir.Properties.Path + "/" + dirName
		}

		return true, nil
	})

	if err != nil {
		if errors.Is(err, appGlobals.ErrInternalServerError) {
			log.Println(fmt.Errorf("rfsCmdModel.go: NewDirectory: %s", err))
			return false, appGlobals.ErrInternalServerError
		}

		return false, err
	}

	return result.(bool), nil
}
