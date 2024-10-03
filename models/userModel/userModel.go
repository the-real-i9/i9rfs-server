package user

import (
	"context"
	"errors"
	"fmt"
	"i9rfs/server/appGlobals"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type User struct {
	Id       string
	Username string
	Password string
}

func New(email, username, password string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := appGlobals.DB.Collection("user").InsertOne(ctx, bson.M{"email": email, "username": username, "password": password})
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: New: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	user := &User{Id: res.InsertedID.(bson.ObjectID).Hex(), Username: username}

	return user, nil
}

func FindOne(uniqueIdent string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	uniqueIdentOid, _ := bson.ObjectIDFromHex(uniqueIdent)

	var resUser struct {
		Id       bson.ObjectID `bson:"_id"`
		Username string
		Password string
	}

	res := appGlobals.DB.Collection("user").FindOne(ctx, bson.M{"$or": bson.A{bson.M{"_id": uniqueIdentOid}, bson.M{"email": uniqueIdent}, bson.M{"username": uniqueIdent}}}, options.FindOne().SetProjection(bson.M{"username": 1, "password": 1}))
	if err := res.Decode(&resUser); err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		log.Println(fmt.Errorf("userModel.go: FindOne: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return &User{Id: resUser.Id.Hex(), Username: resUser.Username, Password: resUser.Password}, nil
}
