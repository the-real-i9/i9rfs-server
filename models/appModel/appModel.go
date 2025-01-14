package appModel

import (
	"context"
	"i9rfs/server/appGlobals"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func AccountExists(ctx context.Context, emailOrUsername string) (bool, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		RETURN EXISTS {
			MATCH (u:User)
			WHERE email = $emailOrUsername OR username = $emailOrUsername
		} AS user_exists
		`,
		map[string]any{
			"emailOrUsername": emailOrUsername,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithReadersRouting(),
	)
	if err != nil {
		log.Println("appModel.go: AccountExists:", err)
		return false, appGlobals.ErrInternalServerError
	}

	exists, _, err := neo4j.GetRecordValue[bool](res.Records[0], "user_exists")
	if err != nil {
		log.Println("appModel.go: AccountExists:", err)
		return false, appGlobals.ErrInternalServerError
	}

	return exists, nil
}
