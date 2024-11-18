package appServices

import (
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"
)

func NewSession(tableName string, sessionData any) (string, error) {
	sessionId, err := helpers.QueryRowField[string](`INSERT INTO $1 (session_data) VALUES ($2) RETURNING session_id`, tableName, sessionData)
	if err != nil {
		log.Println("appServices.go: NewSession:", err)
		return "", appGlobals.ErrInternalServerError
	}

	return *sessionId, nil
}

func RetrieveSession[T any](tableName, sessionId string) (*T, error) {
	sessionData, err := helpers.QueryRowField[T]("SELECT session_data FROM $1 WHERE session_id = $2", tableName, sessionId)
	if err != nil {
		log.Println("appServices.go: RetrieveSession:", err)
		return nil, appGlobals.ErrInternalServerError
	}

	return sessionData, nil
}

func UpdateSession(tableName, sessionId string, sessionData any) error {
	_, err := helpers.QueryRowField[string](`UPDATE $1 SET session_data = $2 WHERE session_id = $3`, tableName, sessionData, sessionId)
	if err != nil {
		log.Println("appServices.go: UpdateSession:", err)
		return appGlobals.ErrInternalServerError
	}

	return nil
}

func EndSession(tableName, sessionId string) {
	go helpers.QueryRowField[bool]("DELETE FROM $1 WHERE session_id = $2", tableName, sessionId)
}
