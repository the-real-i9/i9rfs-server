package user

import (
	"context"
	"fmt"
	"i9rfs/src/appGlobals"
	"i9rfs/src/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func New(ctx context.Context, email, username, password string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		CREATE (u:User { email: $email, username: $username, password: $password })
		
		CREATE (root:UserRoot{ user: $username }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Documents", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Downloads", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Music", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Pictures", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Videos", date_created: $now, date_modified: $now, native: true, starred: false })
		
		CREATE (:UserTrash{ user: $username })
			
		RETURN u { .username } AS new_user
		`,
		map[string]any{
			"email":    email,
			"username": username,
			"password": password,
			"now":      time.Now().UTC(),
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: New: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	newUser, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_user")

	return newUser, nil
}

func AuthFind(ctx context.Context, emailOrUsername string) (map[string]any, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		MATCH (u:User)
		WHERE u.email = $emailOrUsername OR u.username = $emailOrUsername
		RETURN u { .username, .password } AS found_user
		`,
		map[string]any{
			"emailOrUsername": emailOrUsername,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: AuthFind: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, nil
	}

	foundUser, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return foundUser, nil
}

func Exists(ctx context.Context, emailOrUsername string) (bool, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		RETURN EXISTS {
			MATCH (u:User)
			WHERE u.email = $emailOrUsername OR u.username = $emailOrUsername
		} AS user_exists
		`,
		map[string]any{
			"emailOrUsername": emailOrUsername,
		},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithReadersRouting(),
	)
	if err != nil {
		log.Println("userModel.go: Exists:", err)
		return false, fiber.ErrInternalServerError
	}

	exists, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "user_exists")

	return exists, nil
}
