package user

import (
	"context"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type user struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func New(ctx context.Context, email, username, password string) (*user, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		CREATE (u:User { id: randomUUID(), email: $email, username: $username, password: $password })
		RETURN u { .id, .username, .password } AS new_user
		`,
		map[string]any{
			"email":    email,
			"username": username,
			"password": password,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: New: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	recVal, _, err := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_user")
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: New: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	var newUser user

	helpers.MapToStruct(recVal, &newUser)

	return &newUser, nil
}

func FindById(ctx context.Context, userId string) (*user, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		MATCH (u:User { id: $userId })
		RETURN u { .id, .username, .password } AS found_user
		`,
		map[string]any{
			"userId": userId,
		},
		neo4j.EagerResultTransformer,
	)
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindById: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	recVal, _, err := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindById: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	var foundUser user

	helpers.MapToStruct(recVal, &foundUser)

	return &foundUser, nil
}

func FindByEmailOrUsername(ctx context.Context, emailOrUsername string) (*user, error) {
	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		`
		MATCH (u:User)
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

	recVal, _, err := neo4j.GetRecordValue[map[string]any](res.Records[0], "found_user")
	if err != nil {
		log.Println(fmt.Errorf("userModel.go: FindByEmailOrUsername: %s", err))
		return nil, appGlobals.ErrInternalServerError
	}

	var foundUser user

	helpers.MapToStruct(recVal, &foundUser)

	return &foundUser, nil
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
		return false, appGlobals.ErrInternalServerError
	}

	exists, _, err := neo4j.GetRecordValue[bool](res.Records[0], "user_exists")
	if err != nil {
		log.Println("appModel.go: AccountExists:", err)
		return false, appGlobals.ErrInternalServerError
	}

	return exists, nil
}
