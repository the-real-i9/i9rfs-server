package user

import (
	"context"
	"fmt"
	"i9rfs/appGlobals"
	"i9rfs/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func New(ctx context.Context, email, username, password string) (map[string]any, error) {
	res, err := db.Query(
		ctx,
		`
		CREATE (u:User { id: randomUUID(), email: $email, username: $username, password: $password })
		
		CREATE (root:UserRoot{ user: $username }), (:UserTrash{ user: $username }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory" name: "Documents", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory" name: "Downloads", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory" name: "Music", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory" name: "Pictures", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory" name: "Videos", date_created: $now, date_modified: $now, native: true, starred: false })
			
		RETURN u { .id, .username } AS new_user
		`,
		map[string]any{
			"email":    email,
			"username": username,
			"password": password,
			"now":      time.Now(),
		},
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: New: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	newUser, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_user")

	return newUser, nil
}

func FindOne(ctx context.Context, emailOrUsername string) (map[string]any, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		OPTIONAL MATCH (u:User)
		WHERE u.email = $emailOrUsername OR u.username = emailOrUsername
		RETURN u { .id, .username, .password } AS found_user
		`,
		map[string]any{
			"emailOrUsername": emailOrUsername,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindByEmailOrUsername: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	foundUser, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")

	return foundUser, nil
}

func Exists(ctx context.Context, emailOrUsername string) (bool, error) {
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
		return false, fiber.ErrInternalServerError
	}

	exists, _, _ := neo4j.GetRecordValue[bool](res.Records[0], "user_exists")

	return exists, nil
}
