package rfsCmdModel

import (
	"context"
	"i9rfs/models/db"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Ls(ctx context.Context, clientUsername, directoryId string) ([]any, error) {
	var cypher string

	if directoryId == "/" {
		cypher = `
		OPTIONAL MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj)
		WITH obj
		ORDER BY obj.obj_type DESC, obj.name ASC
		RETURN collect(obj { .* }) AS dir_cont
		`
	} else {
		cypher = `
		OPTIONAL MATCH (:UserRoot{ user: $client_username })
		(()-[:HAS_CHILD]->())+
		(:Object{ id: $directory_id })-[:HAS_CHILD]->(obj)
		ORDER BY obj.obj_type DESC, obj.name ASC
		RETURN collect(obj { .* }) AS dir_cont
		`
	}

	res, err := db.Query(
		ctx,
		cypher,
		map[string]any{
			"client_username": clientUsername,
			"directory_id":    directoryId,
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: Ls:", err)
		return nil, fiber.ErrInternalServerError
	}

	dirCont, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "dir_cont")

	return dirCont, nil
}
