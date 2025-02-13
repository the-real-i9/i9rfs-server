package rfsCmdModel

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

func Ls(ctx context.Context, clientUsername, directoryId string) ([]any, error) {
	var cypher string

	if directoryId == "/" {
		cypher = `
		OPTIONAL MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.trashed IS NULL)
		WITH obj, toString(obj.date_created) AS date_created, toString(obj.date_modified) AS date_modified
		ORDER BY obj.obj_type DESC, obj.name ASC
		RETURN collect(obj { .*, date_created, date_modified }) AS dir_cont
		`
	} else {
		cypher = `
		OPTIONAL MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $directory_id })-[:HAS_CHILD]->(obj WHERE obj.trashed IS NULL)
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
		CREATE (root)-[:HAS_CHILD]->(obj:Object{ id: randomUUID(), obj_type: "directory" name: $dir_name, date_created: $now, date_modified: $now })
		RETURN obj { .*, date_created, date_modifed } AS new_dir
		`
	} else {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(parObj:Object{ id: $parent_dir_id })
		CREATE (parObj)-[:HAS_CHILD]->(obj:Object{ id: randomUUID(), obj_type: "directory" name: $dir_name, date_created: $now, date_modified: $now })
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
			"now":             time.Now().UTC(),
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
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)(()-[:HAS_CHILD]->(childObj))*

		WITH obj, childObj, [o IN obj WHERE o.obj_type = "file" | o.id] AS objFileIds, [co IN childObj WHERE co.obj_type = "file" | co.id] AS childObjFileIds

		DETACH DELETE obj, childObj
	
		RETURN objFileIds + childObjFileIds AS file_ids
		`
	} else {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)(()-[:HAS_CHILD]->(childObj))*
			
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
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)

		SET obj.trashed = true, obj.trashed_on = $now

		MATCH (trash:UserTrash{ user: $client_username })
		CREATE (trash)-[:HAS_CHILD]->(obj)
		`
	} else {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)
			
		SET obj.trashed = true, obj.trashed_on = $now

		MATCH (trash:UserTrash{ user: $client_username })
		CREATE (trash)-[:HAS_CHILD]->(obj)
		`
	}

	_, err := db.Query(
		ctx,
		cypher,
		map[string]any{
			"client_username": clientUsername,
			"parent_dir_id":   parentDirectoryId,
			"object_ids":      objectIds,
			"now":             time.Now().UTC(),
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
		OPTIONAL MATCH (:UserTrash{ user: $client_username })-[tr:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)

		DELETE tr

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
		OPTIONAL MATCH (:UserTrash{ user: $client_username })-[:HAS_CHILD]->(obj)

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

func Rename(ctx context.Context, clientUsername, parentDirectoryId, objectId, newName string) (bool, error) {
	var cypher string

	if parentDirectoryId == "/" {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj:Object{ id: $object_id } WHERE obj.native IS NULL)

		SET obj.name = $new_name, obj.date_modified = $now
		`
	} else {
		cypher = `
		MATCH (:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj:Object{ id: $object_id })
			
		SET obj.name = $new_name, obj.date_modified = $now
		`
	}

	_, err := db.Query(
		ctx,
		cypher,
		map[string]any{
			"client_username": clientUsername,
			"parent_dir_id":   parentDirectoryId,
			"object_id":       objectId,
			"now":             time.Now().UTC(),
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: Rename:", err)
		return false, fiber.ErrInternalServerError
	}

	return true, nil
}

func Move(ctx context.Context, clientUsername, fromParentDirectoryId, toParentDirectoryId string, objectIds []string) (bool, error) {
	if fromParentDirectoryId == toParentDirectoryId {
		return false, fiber.NewError(fiber.StatusBadRequest, "attempt to move to the same directory")
	}

	var cypher string

	if fromParentDirectoryId == "/" && toParentDirectoryId != "/" {
		cypher = `
		MATCH (root:UserRoot{ user: $client_username }),
			(root)-[old:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)

		DELETE old

		MATCH (root)-[:HAS_CHILD]->+(toParDir:Object{ id: $to_parent_dir_id })

		SET toParDir.date_modified = $now

		CREATE (toParDir)-[:HAS_CHILD]->(obj)
		`
	} else if fromParentDirectoryId != "/" && toParentDirectoryId == "/" {
		cypher = `
		MATCH (root:UserRoot{ user: $client_username }),
			(root)-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })-[old:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)

		SET fromParDir.date_modified = $now

		DELETE old

		CREATE (root)-[:HAS_CHILD]->(obj)
		`
	} else {
		cypher = `
		MATCH (root:UserRoot{ user: $client_username }),
			(root)-[:HAS_CHILD]->+(toParDir:Object{ id: $to_parent_dir_id }),
			(root)-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })-[old:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)

		SET fromParDir.date_modified = $now, toParDir.date_modified = $now

		DELETE old

		CREATE (toParDir)-[:HAS_CHILD]->(obj)
		`
	}

	_, err := db.Query(
		ctx,
		cypher,
		map[string]any{
			"client_username":    clientUsername,
			"from_parent_dir_id": fromParentDirectoryId,
			"to_parent_dir_id":   toParentDirectoryId,
			"object_ids":         objectIds,
			"now":                time.Now().UTC(),
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: Move:", err)
		return false, fiber.ErrInternalServerError
	}

	return true, nil
}

func Copy(ctx context.Context, clientUsername, fromParentDirectoryId, toParentDirectoryId string, objectIds []string) (bool, []any, error) {
	if fromParentDirectoryId == toParentDirectoryId {
		return false, nil, fiber.NewError(fiber.StatusBadRequest, "attempt to copy to the same directory")
	}

	sess := appGlobals.Neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})

	defer func() {
		if err := sess.Close(ctx); err != nil {
			log.Println("rfsCmdModel.go: Copy: sess.Close:", err)
		}
	}()

	res, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var matchPath string
		var matchIdent string

		if fromParentDirectoryId == "/" {
			matchPath = "(root:UserRoot{ user: $client_username })"
			matchIdent = "root"
		} else {
			matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })"
			matchIdent = "fromParDir"
		}

		_, err := tx.Run(
			ctx,
			fmt.Sprintf(`
			MATCH %s
			MATCH (%s)-[HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)
				((p)-[:HAS_CHILD]->(c))*

			MERGE (cp:Object{ copied_id: p.id })
			ON CREATE
				SET cp += p { .*, id: randomUUID() }
			
			CREATE (cp)-[:HAS_CHILD]->(cc:Object{ copied_id: c.id })
			SET cc += c { .*, id: randomUUID() }
			`,
				// If `p` happens to be a file or an empty folder, `c` will be null
				// But we always have to create child copies`cc`, therefore,
				// the child copies, `cc`, in these cases are considered "bad copies" (with incomplete or no properties),
				// because they have no corresponding copy source,
				// and they will be deleted in the next query
				matchPath, matchIdent,
			),
			map[string]any{
				"client_username":    clientUsername,
				"from_parent_dir_id": fromParentDirectoryId,
				"object_ids":         objectIds,
			},
		)
		if err != nil {
			return nil, err
		}

		if toParentDirectoryId == "/" {
			matchPath = "(root:UserRoot{ user: $client_username })"
			matchIdent = "root"
		} else {
			matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(toParDir:Object{ id: $to_parent_dir_id })"
			matchIdent = "toParDir"
		}

		// the `badObj` are bad copies that will be deleted
		res, err2 := tx.Run(
			ctx,
			fmt.Sprintf(`
			MATCH %s

			MATCH (obj:Object WHERE obj.copied_id IN $object_ids)-[:HAS_CHILD]->+(badObj WHERE badObj.copied_id IS NULL)

			DETACH DELETE badObj

			CREATE (%s)->[:HAS_CHILD]->(obj)

			WITH obj
			MATCH (obj)-[:HAS_CHILD]->*(cobj)

			WITH [o IN obj WHERE o.obj_type = "file" | o { .copied_id, .id }] AS objFileCopyIdMaps, 
				[co IN cobj WHERE co IS NOT NULL AND co.obj_type = "file" | co { .copied_id, .id }] AS cobjFileCopyIdMaps

			RETURN objFileCopyIdMaps + cobjFileCopyIdMaps AS file_copy_id_maps
			`, matchPath, matchIdent),
			map[string]any{
				"client_username":  clientUsername,
				"to_parent_dir_id": toParentDirectoryId,
				"object_ids":       objectIds,
			},
		)
		if err2 != nil {
			return nil, err2
		}

		fileCopyIdMaps, _, _ := neo4j.GetRecordValue[[]any](res.Record(), "file_copy_id_maps")

		return fileCopyIdMaps, nil
	})
	if err != nil {
		log.Println("rfsCmdModel.go: Copy:", err)
		return false, nil, fiber.ErrInternalServerError
	}

	return true, res.([]any), nil
}
