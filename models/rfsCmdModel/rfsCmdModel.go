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

	// In transaction
	result, err := helpers.MultiOpQuery(db.Client(), func(ctx context.Context) (any, error) {
		var parentDir struct {
			Oid        bson.ObjectID `bson:"_id"`
			Properties struct {
				Path string `bson:"path"`
			} `bson:"properties"`
		}

		// retrieve the parent directory's oid and path from the database
		// the parent directory is one whose path is parentDirPath
		// if the parentDirPath is "/", we won't find a parent directory, and parentDir will have nil values,
		// this new directory, hence, will have no parent i.e it will conceptually be in the root directory ("/newDir")
		err := db.Collection("directory").FindOne(ctx, bson.M{"properties.path": bson.M{"$eq": parentDirPath}}, options.FindOne().SetProjection(bson.M{"_id": 1, "properties.path": 1})).Decode(&parentDir)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return false, err
		}

		// we want every new directory in the tree to be created at the same "time"
		dirDate := time.Now().Format("2 January 2006 3:04:05 PM")

		// since the user is able to specify a directory path separated by "/" to create a directory (degenerate) tree
		// each directory in the (degenerate) tree will be the parent of the next
		// the first directory in the (degenerate) tree will have parentDir above, as its parent
		for _, dirName := range newDirTree {
			dirName := strings.Trim(dirName, "\"")
			var dir struct {
				Oid        bson.ObjectID `bson:"_id"`
				Properties struct {
					Path string `bson:"path"`
				} `bson:"properties"`
			}

			// check if a directory along the tree path already exists
			err := db.Collection("directory").FindOne(ctx, bson.M{"properties.path": bson.M{"$eq": parentDir.Properties.Path + "/" + dirName}}, options.FindOne().SetProjection(bson.M{"_id": 1, "properties.path": 1})).Decode(&dir)
			if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
				return false, err
			}

			// if a directory along the tree path already exists
			// rather than raising an error, we just go ahead and use it
			// thus we make it our parentDir for the next directory in the tree
			// and skip creating a duplicate
			if dir.Properties.Path != "" {
				parentDir.Oid = dir.Oid
				parentDir.Properties.Path = dir.Properties.Path
				continue
			}

			// if a directory along the tree path does not already exists we create it
			dirDoc := bson.M{
				"owner_user_id": userOid,
				"properties": bson.M{
					"name":          dirName,
					"path":          parentDir.Properties.Path + "/" + dirName,
					"date_modified": dirDate,
					"date_created":  dirDate,
				},
			}

			// meanwhile, if we have no parent directory,
			// (i.e. our starting parentDirPath is "/", and, of course, our parentDir has nil values)
			// this new directory is going to be directly in the root (i.e. "/newDir")
			// and we won't give it the parent_directory_id in the database
			// (the property is, therefore, optional in the validation schema)

			// otherwhise, we give this new directory as a child to
			//the previous directory in the tree, which is currently the parent
			if parentDir.Properties.Path != "" {
				dirDoc["parent_directory_id"] = parentDir.Oid
			}

			res, err := db.Collection("directory").InsertOne(ctx, dirDoc)
			if err != nil {
				return nil, err
			}

			// having created the directory
			// we make it our parentDir for the next directory in the tree
			parentDir.Oid = res.InsertedID.(bson.ObjectID)
			parentDir.Properties.Path = parentDir.Properties.Path + "/" + dirName

			// ...going to create the next directory in the tree (if there's more)...
		}

		// new directory operation is successful
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
