package user

import (
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"
)

type user struct {
	Id       string
	Username string
	Password string
}

func New(email, username, password string) (*user, error) {
	newUser, err := helpers.QueryRowType[user]("SELECT * FROM new_user($1, $2, $3)", email, username, password)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: New: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return newUser, nil
}

func FindOne(userId string) (*user, error) {
	resUser, err := helpers.QueryRowType[user]("SELECT * FROM get_user($1)", userId)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindOne: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return resUser, nil
}
