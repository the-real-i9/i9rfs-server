package user

import (
	"fmt"
	"i9rfs/server/globals"
	"i9rfs/server/helpers"
	"log"
)

type User struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}

func New(email, username, password string) (*User, error) {

	user, err := helpers.QueryRowType[User]("SELECT * FROM new_user($1, $2, $3)", email, username, password)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: NewUser: %s", err))
		return nil, globals.ErrInternalServerError
	}

	return user, nil
}

func FindOne(uniqueId string) (*User, error) {

	user, err := helpers.QueryRowType[User]("SELECT * FROM get_user($1)", uniqueId)

	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindOne: %s", err))
		return nil, globals.ErrInternalServerError
	}

	return user, nil
}
