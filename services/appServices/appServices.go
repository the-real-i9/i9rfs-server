package appServices

import (
	"context"
	"encoding/json"
	"i9rfs/server/appGlobals"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func NewSession(ctx context.Context, tableName string, sessionData any) (string, error) {
	data, err := json.Marshal(sessionData)
	if err != nil {
		log.Println("appServices.go: NewSession:", err)
		return "", appGlobals.ErrInternalServerError
	}

	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		CREATE (sess:$($tableName) { sessionId: randomUUID(), sessionData: $sessionData })
		RETURN sess.sessionId AS session_id
		`,
		map[string]any{
			"tableName":   tableName,
			"sessionData": string(data),
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		log.Println("appServices.go: NewSession:", err)
		return "", appGlobals.ErrInternalServerError
	}

	sid, _, err := neo4j.GetRecordValue[string](res.Records[0], "session_id")
	if err != nil {
		log.Println("appServices.go: NewSession:", err)
		return "", appGlobals.ErrInternalServerError
	}

	return sid, nil
}

func RetrieveSession[T any](ctx context.Context, tableName, sessionId string) (*T, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		MATCH (sess:$($tableName) { sessionId: $sessionId })
		RETURN sess.sessionData AS session_data
		`,
		map[string]any{
			"tableName": tableName,
			"sessionId": sessionId,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		log.Println("appServices.go: RetrieveSession:", err)
		return nil, appGlobals.ErrInternalServerError
	}

	sdata, _, err := neo4j.GetRecordValue[string](res.Records[0], "session_data")

	var data T

	jerr := json.Unmarshal([]byte(sdata), &data)
	if jerr != nil {
		log.Println("appServices.go: RetrieveSession:", jerr)
		return nil, appGlobals.ErrInternalServerError
	}

	return &data, nil
}

func UpdateSession(ctx context.Context, tableName, sessionId string, sessionData any) error {
	data, jerr := json.Marshal(sessionData)
	if jerr != nil {
		log.Println("appServices.go: UpdateSession:", jerr)
		return appGlobals.ErrInternalServerError
	}

	_, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		MATCH (sess:$($tableName) { sessionId: $sessionId })
		SET sess.sessionData = $sessionData
		`,
		map[string]any{
			"tableName":   tableName,
			"sessionId":   sessionId,
			"sessionData": string(data),
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		log.Println("appServices.go: UpdateSession:", err)
		return appGlobals.ErrInternalServerError
	}

	return nil
}

func EndSession(ctx context.Context, tableName, sessionId string) {
	go neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		MATCH (sess:$($tableName) { sessionId: $sessionId })
		DELETE sess
		`,
		map[string]any{
			"tableName": tableName,
			"sessionId": sessionId,
		},
		neo4j.EagerResultTransformer,
	)
}
