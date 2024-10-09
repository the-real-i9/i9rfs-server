package appModel

import (
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"
)

func AccountExists(emailOrUsername string) (bool, error) {
	exist, err := helpers.QueryRowField[bool]("SELECT exist FROM account_exists($1)", emailOrUsername)

	if err != nil {
		log.Println("appModel.go: AccountExists:", err)
		return false, appGlobals.ErrInternalServerError
	}

	return *exist, nil
}
