package user

import (
	"context"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

func New(email, username, password string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := appGlobals.DB.Collection("user").InsertOne(ctx, bson.M{"email": email, "username": username, "password": password})
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: NewUser: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	user := &User{Id: res.InsertedID.(bson.ObjectID).Hex(), Username: username}

	return user, nil
}

func FindOne(uniqueId string) (*User, error) {

	user, err := helpers.QueryRowType[User]("SELECT * FROM get_user($1)", uniqueId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindOne: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return user, nil
}
