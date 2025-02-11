package rfsCmdModel

import (
	"context"
	"i9rfs/models/db"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Ls(ctx context.Context, clientUsername, directoryId string) ([]any, error) {
	var cypher string

	if directoryId == "/" {
		cypher = `
		OPTIONAL MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj:Object WHERE obj.trashed <> true)
		WITH obj, toString(obj.date_created) AS date_created, toString(obj.date_modified) AS date_modified
		ORDER BY obj.obj_type DESC, obj.name ASC
		RETURN collect(obj { .*, date_created, date_modified }) AS dir_cont
		`
	} else {
		cypher = `
		OPTIONAL MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $directory_id })-[:HAS_CHILD]->(obj:Object WHERE obj.trashed <> true)
		WITH obj, toString(obj.date_created) AS date_created, toString(obj.date_modified) AS date_modified
		ORDER BY obj.obj_type DESC, obj.name ASC
		RETURN collect(obj { .*, date_created, date_modified }) AS dir_cont
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

func Mkdir(ctx context.Context, clientUsername, parentDirectoryId, directoryName string) (map[string]any, error) {
	var cypher string

	if parentDirectoryId == "/" {
		cypher = `
		MATCH (root:UserRoot{ user: $client_username })
		CREATE (root)-[:HAS_CHILD]->(obj:Object{ id: randomUUID(), obj_type: "directory" name: $dir_name, date_created: $now, date_modified: $now, native: false, starred: false })
		RETURN obj { .*, date_created, date_modifed } AS new_dir
		`
	} else {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(parObj:Object{ id: $parent_dir_id })
		CREATE (parObj)-[:HAS_CHILD]->(obj:Object{ id: randomUUID(), obj_type: "directory" name: $dir_name, date_created: $now, date_modified: $now, native: false, starred: false })
		RETURN obj { .*, date_created, date_modifed } AS new_dir
		`
	}

	res, err := db.Query(
		ctx,
		cypher,
		map[string]any{
			"client_username": clientUsername,
			"parent_dir_id":   parentDirectoryId,
			"dir_name":        directoryName,
			"now":             time.Now(),
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: Mkdir:", err)
		return nil, fiber.ErrInternalServerError
	}

	newDir, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_dir")

	return newDir, nil
}

func Del(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, []any, error) {
	var cypher string

	if parentDirectoryId == "/" {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj:Object WHERE obj.id IN $object_ids)(()-[:HAS_CHILD]->(childObj))*

		WITH obj, childObj, [o IN obj WHERE o.obj_type = "file" | o.id] AS objFileIds, [co IN childObj WHERE co.obj_type = "file" | co.id] AS childObjFileIds

		DETACH DELETE obj, childObj
	
		RETURN objFileIds + childObjFileIds AS file_ids
		`
	} else {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj:Object WHERE obj.id IN $object_ids)(()-[:HAS_CHILD]->(childObj))*
			
		WITH obj, childObj, [o IN obj WHERE o.obj_type = "file" | o.id] AS objFileIds, [co IN childObj WHERE co.obj_type = "file" | co.id] AS childObjFileIds

		DETACH DELETE obj, childObj
	
		RETURN objFileIds + childObjFileIds AS file_ids
		`
	}

	res, err := db.Query(
		ctx,
		cypher,
		map[string]any{
			"client_username": clientUsername,
			"parent_dir_id":   parentDirectoryId,
			"object_ids":      objectIds,
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: Del:", err)
		return false, nil, fiber.ErrInternalServerError
	}

	fileIds, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "file_ids")

	return true, fileIds, nil
}

func Trash(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, error) {
	var cypher string

	if parentDirectoryId == "/" {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj:Object WHERE obj.id IN $object_ids)

		SET obj.trashed = true, obj.trashed_on = $now
		`
	} else {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj:Object WHERE obj.id IN $object_ids)
			
		SET obj.trashed = true, obj.trashed_on = $now
		`
	}

	_, err := db.Query(
		ctx,
		cypher,
		map[string]any{
			"client_username": clientUsername,
			"parent_dir_id":   parentDirectoryId,
			"object_ids":      objectIds,
			"now":             time.Now(),
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: Trash:", err)
		return false, fiber.ErrInternalServerError
	}

	return true, nil
}

func Restore(ctx context.Context, clientUsername string, objectIds []string) (bool, error) {
	_, err := db.Query(
		ctx,
		`
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(obj:Object WHERE obj.id IN $object_ids)

		SET obj.trashed = null, obj.trashed_on = null
		`,
		map[string]any{
			"client_username": clientUsername,
			"object_ids":      objectIds,
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: Restore:", err)
		return false, fiber.ErrInternalServerError
	}

	return true, nil
}

func ShowTrash(ctx context.Context, clientUsername string) ([]any, error) {
	res, err := db.Query(
		ctx,
		`
		OPTIONAL MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(obj:Object WHERE obj.trashed = true)

		WITH obj, toString(obj.date_created) AS date_created, toString(obj.date_modified) AS date_modified, toString(obj.trashed_on) AS trashed_on
		ORDER BY obj.obj_type DESC, obj.name ASC
		RETURN collect(obj { .*, date_created, date_modified, trashed_on }) AS trash_cont
		`,
		map[string]any{
			"client_username": clientUsername,
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: ShowTrash:", err)
		return nil, fiber.ErrInternalServerError
	}

	trashCont, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "trash_cont")

	return trashCont, nil
}
