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

func FindById(userId string) (*user, error) {
	resUser, err := helpers.QueryRowType[user]("SELECT * FROM find_user_by_id($1)", userId)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindById: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return resUser, nil
}

func FindByEmailOrUsername(emailOrUsername string) (*user, error) {
	resUser, err := helpers.QueryRowType[user]("SELECT * FROM find_user_by_email_or_username($1)", emailOrUsername)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindByEmailOrUsername: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	return resUser, nil
}
