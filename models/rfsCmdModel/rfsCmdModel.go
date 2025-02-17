package rfsCmdModel

import (
	"context"
	"fmt"
	"i9rfs/appGlobals"
	"i9rfs/models/db"
	"log"
	"maps"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Ls(ctx context.Context, clientUsername, directoryId string) ([]any, error) {
	var matchPath string

	if directoryId == "/" {
		matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.trashed IS NULL)"
	} else {
		matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $directory_id })-[:HAS_CHILD]->(obj WHERE obj.trashed IS NULL)"
	}

	res, err := db.Query(
		ctx,
		fmt.Sprintf(`
			MATCH %s
			WITH obj, toString(obj.date_created) AS date_created, toString(obj.date_modified) AS date_modified
			ORDER BY obj.obj_type DESC, obj.name ASC
			RETURN collect(obj { .*, date_created, date_modified }) AS dir_cont
		`, matchPath),
		map[string]any{
			"client_username": clientUsername,
			"directory_id":    directoryId,
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: Ls:", err)
		return nil, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'directoryId'")
	}

	dirCont, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "dir_cont")

	return dirCont, nil
}

func Mkdir(ctx context.Context, clientUsername, parentDirectoryId, directoryName string) (map[string]any, error) {
	var matchPath string
	var matchIdent string

	if parentDirectoryId == "/" {
		matchPath = "(root:UserRoot{ user: $client_username })"
		matchIdent = "root"
	} else {
		matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(parObj:Object{ id: $parent_dir_id })"
		matchIdent = "parObj"
	}

	res, err := db.Query(
		ctx,
		fmt.Sprintf(`
			MATCH %s
			CREATE (%s)-[:HAS_CHILD]->(obj:Object{ id: randomUUID(), obj_type: "directory", name: $dir_name, date_created: $now, date_modified: $now })
			
			WITH obj, toString(obj.date_created) AS date_created, toString(obj.date_modified) AS date_modified
			RETURN obj { .*, date_created, date_modified } AS new_dir
		`, matchPath, matchIdent),
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

	if len(res.Records) == 0 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'parentDirectoryId'")
	}

	newDir, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_dir")

	return newDir, nil
}

func Del(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, []any, error) {

	var matchPath string

	if parentDirectoryId == "/" {
		matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)(()-[:HAS_CHILD]->(childObjs))*"
	} else {
		matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)(()-[:HAS_CHILD]->(childObjs))*"
	}

	res, err := db.Query(
		ctx,
		fmt.Sprintf(`
			MATCH %s
			
			WITH obj, childObjs,
				[o IN collect(obj) WHERE o.obj_type = "file" | o.id] AS objFileIds,
				[co IN childObjs WHERE co.obj_type = "file" | co.id] AS childObjFileIds

			DETACH DELETE obj

			WITH objFileIds, childObjFileIds, childObjs

			UNWIND (childObjs + [null]) AS cObj
			DETACH DELETE cObj
			WITH objFileIds, childObjFileIds

			RETURN objFileIds + childObjFileIds AS file_ids
		`, matchPath),
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

	if len(res.Records) == 0 {
		return false, nil, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'parentDirectoryId', and valid 'objectIds' in the directory | no value in 'objectIds' is a native directory")
	}

	fileIds, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "file_ids")

	return true, fileIds, nil
}

func Trash(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, error) {
	var matchPath string

	if parentDirectoryId == "/" {
		matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)"
	} else {
		matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)"
	}

	res, err := db.Query(
		ctx,
		fmt.Sprintf(`
			MATCH %s

			SET obj.trashed = true, obj.trashed_on = $now

			WITH obj

			MATCH (trash:UserTrash{ user: $client_username })
			CREATE (trash)-[:HAS_CHILD]->(obj)

			RETURN true AS workdone
		`, matchPath),
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

	if len(res.Records) == 0 {
		return false, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'parentDirectoryId', and valid 'objectIds' in the directory | no value in 'objectIds' is a native directory")
	}

	return true, nil
}

func Restore(ctx context.Context, clientUsername string, objectIds []string) (bool, error) {
	res, err := db.Query(
		ctx,
		`
		MATCH (:UserTrash{ user: $client_username })-[tr:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)

		DELETE tr

		SET obj.trashed = null, obj.trashed_on = null

		RETURN true AS workdone
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

	if len(res.Records) == 0 {
		return false, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying valid 'objectIds' in Trash")
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
	var matchPath string

	if parentDirectoryId == "/" {
		matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj:Object{ id: $object_id } WHERE obj.native IS NULL AND obj.trashed IS NULL)"
	} else {
		matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj:Object{ id: $object_id } WHERE obj.trashed IS NULL)"
	}

	res, err := db.Query(
		ctx,
		fmt.Sprintf(`
			MATCH %s

			SET obj.name = $new_name, obj.date_modified = $now

			RETURN true AS workdone
		`, matchPath),
		map[string]any{
			"client_username": clientUsername,
			"parent_dir_id":   parentDirectoryId,
			"object_id":       objectId,
			"new_name":        newName,
			"now":             time.Now().UTC(),
		},
	)
	if err != nil {
		log.Println("rfsCmdModel.go: Rename:", err)
		return false, fiber.ErrInternalServerError
	}

	if len(res.Records) == 0 {
		return false, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'parentDirectoryId', and a valid 'objectId' in the directory | 'objectId' is not a native directory and not in Trash")
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
			(root)-[old:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL),
			(root)-[:HAS_CHILD]->+(toParDir:Object{ id: $to_parent_dir_id })

		DELETE old

		WITH root, obj
		SET toParDir.date_modified = $now
		CREATE (toParDir)-[:HAS_CHILD]->(obj)

		RETURN true AS workdone
		`
	} else if fromParentDirectoryId != "/" && toParentDirectoryId == "/" {
		cypher = `
		MATCH (root:UserRoot{ user: $client_username }),
			(root)-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })-[old:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)

		SET fromParDir.date_modified = $now

		DELETE old

		WITH root, obj
		CREATE (root)-[:HAS_CHILD]->(obj)

		RETURN true AS workdone
		`
	} else {
		cypher = `
		MATCH (root:UserRoot{ user: $client_username }),
			(root)-[:HAS_CHILD]->+(toParDir:Object{ id: $to_parent_dir_id }),
			(root)-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })-[old:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)

		SET fromParDir.date_modified = $now, toParDir.date_modified = $now

		DELETE old

		WITH toParDir, obj
		CREATE (toParDir)-[:HAS_CHILD]->(obj)

		RETURN true AS workdone
		`
	}

	res, err := db.Query(
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

	if len(res.Records) == 0 {
		return false, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'fromParentDirectoryId', valid 'objectIds' in the directory, and a valid 'toParentDirectoryId' | no value in 'objectIds' is a native directory")
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
		var (
			matchPath  string
			matchIdent string

			res neo4j.ResultWithContext
			err error
			now = time.Now().UTC()
		)

		if fromParentDirectoryId == "/" {
			matchPath = "(root:UserRoot{ user: $client_username })"
			matchIdent = "root"
		} else {
			matchPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })"
			matchIdent = "fromParDir"
		}

		res, err = tx.Run(
			ctx,
			fmt.Sprintf(`
			MATCH %s
			MATCH (%s)-[HAS_CHILD]->(obj WHERE obj.id IN $object_ids)
				((parents)-[:HAS_CHILD]->(children))*

			RETURN [p IN parents | p.id] AS parent_ids, [c IN children | c.id] AS children_ids
			`, matchPath, matchIdent),
			map[string]any{
				"client_username":    clientUsername,
				"from_parent_dir_id": fromParentDirectoryId,
				"object_ids":         objectIds,
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return false, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'fromParentDirectoryId', valid 'objectIds' in the directory, and a valid 'toParentDirectoryId'")
		}

		recMap := make(map[string]any, 2)

		maps.Copy(res.Record().AsMap(), recMap)

		parentIds := recMap["parent_ids"].([]string)
		childrenIds := recMap["children_ids"].([]string)

		parentIdsLen := len(parentIds)

		if parentIdsLen != len(childrenIds) {
			return nil, fmt.Errorf("you have a problem here: parLen and chiLen not equal")
		}

		parChildList := make([][]any, parentIdsLen)

		for i := 0; i < parentIdsLen; i++ {
			parChildList[i] = []any{parentIds[i], childrenIds[i]}
		}

		_, err = tx.Run(
			ctx,
			fmt.Sprintf(`
			MATCH %s
			UNWIND $par_child_list AS par_child
			CALL (%[1]s, par_child) {
				MATCH (%[1]s)-[:HAS_CHILD]->(par:Object{ id: par_child[0] })

				MERGE (pc:Object{ copied_id: par.id })
				ON CREATE
					SET pc += par { .*, id: randomUUID(), native: null, date_created: $now, date_modified: $now }
	
				OPTIONAL MATCH (%[1]s)-[:HAS_CHILD]->(chi:Object{ id: par_child[1] })

				CREATE (pc)-[:HAS_CHILD]->(cc:Object{ copied_id: chi.id })
				SET cc += chi { .*, id: randomUUID(), date_created: $now, date_modified: $now }
			}
			`, matchPath, matchIdent,
			// If `par` happens to be a file or an empty folder, `chi` will be null
			// But we always have to create child copies`cc`, therefore,
			// the child copies, `cc`, in these cases are considered "bad copies" (with incomplete or no properties),
			// because they have no corresponding copy source,
			// and they will be deleted in the next query
			),
			map[string]any{
				"par_child_list":     parChildList,
				"client_username":    clientUsername,
				"from_parent_dir_id": fromParentDirectoryId,
				"now":                now,
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
		res, err = tx.Run(
			ctx,
			fmt.Sprintf(`
			MATCH %s

			MATCH (obj:Object WHERE obj.copied_id IN $object_ids)-[:HAS_CHILD]->+(badObj WHERE badObj.copied_id IS NULL)

			DETACH DELETE badObj

			WITH %[1]s, obj

			CREATE (%[1]s)->[:HAS_CHILD]->(obj)
			
			SET %[1]s.date_modified = $now

			WITH obj
			MATCH (obj)-[:HAS_CHILD]->*(cobjs)

			WITH obj, cobjs, 
				[o IN collect(obj) WHERE o.obj_type = "file" | o { .copied_id, copy_id: o.id }] AS objFileCopyIdMaps,
				[co IN cobjs WHERE co.obj_type = "file" | co { .copied_id, copy_id: co.id }] AS cobjFileCopyIdMaps

			SET obj.copied_id = null

			UNWIND (cobjs + [null]) AS cobj
			SET cobj.copied_id = null

			WITH objFileCopyIdMaps, cobjFileCopyIdMaps

			RETURN objFileCopyIdMaps + cobjFileCopyIdMaps AS file_copy_id_maps
			`, matchPath, matchIdent),
			map[string]any{
				"client_username":  clientUsername,
				"to_parent_dir_id": toParentDirectoryId,
				"object_ids":       objectIds,
				"now":              time.Now().UTC(),
			},
		)
		if err != nil {
			return nil, err
		}

		if res.Record() == nil {
			return false, fiber.NewError(fiber.StatusBadRequest, "logical error! check that: you're specifying a valid 'fromParentDirectoryId', valid 'objectIds' in the directory, and a valid 'toParentDirectoryId'")
		}

		fileCopyIdMaps, _, _ := neo4j.GetRecordValue[[]any](res.Record(), "file_copy_id_maps")

		return fileCopyIdMaps, nil
	})
	if err != nil {
		if fe, ok := res.(*fiber.Error); ok {
			return false, nil, fe
		}

		log.Println("rfsCmdModel.go: Copy:", err)
		return false, nil, fiber.ErrInternalServerError
	}

	return true, res.([]any), nil
}
