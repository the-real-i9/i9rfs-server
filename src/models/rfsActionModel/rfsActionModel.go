package rfsActionModel

import (
	"context"
	"fmt"
	"i9rfs/src/appGlobals"
	"i9rfs/src/models/db"
	"log"
	"maps"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func Ls(ctx context.Context, clientUsername, directoryId string) ([]any, error) {
	var matchFromPath string

	if directoryId == "/" {
		matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.trashed IS NULL)"
	} else {
		matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $directory_id })-[:HAS_CHILD]->(obj WHERE obj.trashed IS NULL)"
	}

	res, err := db.Query(
		ctx,
		fmt.Sprintf(`
			MATCH %s
			WITH obj, toString(obj.date_created) AS date_created, toString(obj.date_modified) AS date_modified
			ORDER BY obj.obj_type DESC, obj.name ASC
			RETURN collect(obj { .*, date_created, date_modified }) AS dir_cont
		`, matchFromPath),
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
		return nil, nil
	}

	dirCont, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "dir_cont")

	return dirCont, nil
}

func Mkdir(ctx context.Context, clientUsername, parentDirectoryId, directoryName string) (map[string]any, error) {
	var matchFromPath string
	var matchFromIdent string

	if parentDirectoryId == "/" {
		matchFromPath = "(root:UserRoot{ user: $client_username })"
		matchFromIdent = "root"
	} else {
		matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(parObj:Object{ id: $parent_dir_id })"
		matchFromIdent = "parObj"
	}

	res, err := db.Query(
		ctx,
		fmt.Sprintf(`
			MATCH %s
			CREATE (%s)-[:HAS_CHILD]->(obj:Object{ id: randomUUID(), obj_type: "directory", name: $dir_name, date_created: $now, date_modified: $now })
			
			WITH obj, toString(obj.date_created) AS date_created, toString(obj.date_modified) AS date_modified
			RETURN obj { .*, date_created, date_modified } AS new_dir
		`, matchFromPath, matchFromIdent),
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
		return nil, nil
	}

	newDir, _, _ := neo4j.GetRecordValue[map[string]any](res.Records[0], "new_dir")

	return newDir, nil
}

func Del(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, []any, error) {

	var matchFromPath string

	if parentDirectoryId == "/" {
		matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)(()-[:HAS_CHILD]->(childObjs))*"
	} else {
		matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)(()-[:HAS_CHILD]->(childObjs))*"
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
		`, matchFromPath),
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
		return false, nil, nil
	}

	fileIds, _, _ := neo4j.GetRecordValue[[]any](res.Records[0], "file_ids")

	return true, fileIds, nil
}

func Trash(ctx context.Context, clientUsername, parentDirectoryId string, objectIds []string) (bool, error) {
	var matchFromPath string

	if parentDirectoryId == "/" {
		matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)"
	} else {
		matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)"
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
		`, matchFromPath),
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
		return false, nil
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
		return false, nil
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
	var matchFromPath string

	if parentDirectoryId == "/" {
		matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj:Object{ id: $object_id } WHERE obj.native IS NULL AND obj.trashed IS NULL)"
	} else {
		matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj:Object{ id: $object_id } WHERE obj.trashed IS NULL)"
	}

	res, err := db.Query(
		ctx,
		fmt.Sprintf(`
			MATCH %s

			SET obj.name = $new_name, obj.date_modified = $now

			RETURN true AS workdone
		`, matchFromPath),
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
		return false, nil
	}

	return true, nil
}

func Move(ctx context.Context, clientUsername, fromParentDirectoryId, toParentDirectoryId string, objectIds []string) (bool, error) {
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
		return false, nil
	}

	return true, nil
}

func Copy(ctx context.Context, clientUsername, fromParentDirectoryId, toParentDirectoryId, objectId string) ([]any, error) {
	sess := appGlobals.Neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})

	defer func() {
		if err := sess.Close(ctx); err != nil {
			log.Println("rfsCmdModel.go: Copy: sess.Close:", err)
		}
	}()

	res, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		var (
			matchFromPath, matchFromIdent string
			matchToPath, matchToIdent     string

			now = time.Now().UTC()
		)

		if fromParentDirectoryId == "/" {
			matchFromPath = "(root:UserRoot{ user: $client_username })"
			matchFromIdent = "root"
		} else {
			matchFromPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })"
			matchFromIdent = "fromParDir"
		}

		if toParentDirectoryId == "/" {
			matchToPath = "(root:UserRoot{ user: $client_username })"
			matchToIdent = "root"
		} else {
			matchToPath = "(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(toParDir:Object{ id: $to_parent_dir_id })"
			matchToIdent = "toParDir"
		}

		var objectHasChildren bool

		{
			res, err := tx.Run(
				ctx,
				fmt.Sprintf(`
					MATCH %[1]s

					RETURN EXISTS { (%[2]s)-[:HAS_CHILD]->(:Object{ id: $object_id })-[:HAS_CHILD]->() } AS object_has_children
				`, matchFromPath, matchFromIdent),
				map[string]any{
					"client_username":    clientUsername,
					"from_parent_dir_id": fromParentDirectoryId,
					"object_id":          objectId,
				},
			)
			if err != nil {
				return nil, err
			}

			if !res.Next(ctx) {
				return nil, nil
			}

			objectHasChildren, _, _ = neo4j.GetRecordValue[bool](res.Record(), "object_has_children")
		}

		var fileCopyIdMaps []any

		if objectHasChildren {
			res, err := tx.Run(
				ctx,
				fmt.Sprintf(`
				MATCH %[1]s
				MATCH (%[2]s)-[:HAS_CHILD]->(obj:Object{ id: $object_id })
					((parents)-[:HAS_CHILD]->(children))+
	
				RETURN [p IN parents | p.id] AS parent_ids, [c IN children | c.id] AS children_ids
				`, matchFromPath, matchFromIdent),
				map[string]any{
					"client_username":    clientUsername,
					"from_parent_dir_id": fromParentDirectoryId,
					"object_id":          objectId,
				},
			)
			if err != nil {
				return nil, err
			}

			if !res.Next(ctx) {
				return nil, nil
			}

			recMap := make(map[string]any)

			recs, _ := res.Collect(ctx)

			// last record contains the full parents and children
			maps.Copy(recMap, recs[len(recs)-1].AsMap())

			parentIds := recMap["parent_ids"].([]any)
			childrenIds := recMap["children_ids"].([]any)

			parentIdsLen := len(parentIds)

			parentIdToChildId := make([][]any, parentIdsLen)

			for i := range parentIdsLen {
				parentIdToChildId[i] = []any{parentIds[i], childrenIds[i]}
			}

			{
				_, err := tx.Run(
					ctx,
					fmt.Sprintf(`
					MATCH %[1]s

					UNWIND $par_id_to_child_id AS par_id_0_chi_id_1
					CALL (%[2]s, par_id_0_chi_id_1) {
						MATCH (%[2]s)-[:HAS_CHILD]->+(par:Object{ id: par_id_0_chi_id_1[0] })

						MERGE (parentCopy:Object{ copied_id: par.id })
						ON CREATE
							SET parentCopy += par { .*, id: randomUUID(), native: null, date_created: $now, date_modified: $now }
			
						WITH par, par_id_0_chi_id_1, parentCopy

						MATCH (par)-[:HAS_CHILD]->(chi:Object{ id: par_id_0_chi_id_1[1] })

						CREATE (parentCopy)-[:HAS_CHILD]->(childCopy:Object{ copied_id: chi.id })
						SET childCopy += chi { .*, id: randomUUID(), date_created: $now, date_modified: $now }
					}
					`, matchFromPath, matchFromIdent,
					),
					map[string]any{
						"par_id_to_child_id": parentIdToChildId,
						"client_username":    clientUsername,
						"from_parent_dir_id": fromParentDirectoryId,
						"now":                now,
					},
				)
				if err != nil {
					return nil, err
				}
			}

			{
				res, err := tx.Run(
					ctx,
					fmt.Sprintf(`
					MATCH %[1]s

					MATCH (obj:Object { copied_id: $object_id })

					CREATE (%[2]s)-[:HAS_CHILD]->(obj)

					WITH %[2]s, obj
					
					MATCH (obj)-[:HAS_CHILD]->*(cobj)

					WITH %[2]s, obj, cobj, 
						[o IN collect(obj) WHERE o.obj_type = "file" | o { .copied_id, copy_id: o.id }] AS objFileCopyIdMaps,
						[co IN collect(cobj) WHERE co.obj_type = "file" | co { .copied_id, copy_id: co.id }] AS cobjFileCopyIdMaps

					SET obj.copied_id = null,
						cobj.copied_id = null
					
					SET %[2]s.date_modified = $now

					RETURN objFileCopyIdMaps + cobjFileCopyIdMaps AS file_copy_id_maps
					`, matchToPath, matchToIdent),
					map[string]any{
						"client_username":  clientUsername,
						"to_parent_dir_id": toParentDirectoryId,
						"object_id":        objectId,
						"now":              now,
					},
				)
				if err != nil {
					return nil, err
				}

				if !res.Next(ctx) {
					return nil, nil
				}

				fileCopyIdMaps, _, _ = neo4j.GetRecordValue[[]any](res.Record(), "file_copy_id_maps")
			}
		} else {
			res, err := tx.Run(
				ctx,
				fmt.Sprintf(`
				MATCH %[1]s
				MATCH %[3]s

				MATCH (%[2]s)-[:HAS_CHILD]->(obj:Object{ id: $object_id })

				CREATE (%[4]s)-[:HAS_CHILD]->(objCopy:Object)
				SET objCopy += obj { .*, id: randomUUID(), native: null, date_created: $now, date_modified: $now }

				SET %[4]s.date_modified = $now

				RETURN 
					CASE obj.obj_type 
						WHEN = "file" THEN [{ copied_id: $object_id, copy_id: objCopy }]
						ELSE []
					END AS file_copy_id_maps
				`, matchFromPath, matchFromIdent, matchToPath, matchToIdent),
				map[string]any{
					"client_username":    clientUsername,
					"from_parent_dir_id": fromParentDirectoryId,
					"to_parent_dir_id":   toParentDirectoryId,
					"object_id":          objectId,
					"now":                now,
				},
			)
			if err != nil {
				return nil, err
			}

			if !res.Next(ctx) {
				return nil, nil
			}

			fileCopyIdMaps, _, _ = neo4j.GetRecordValue[[]any](res.Record(), "file_copy_id_maps")
		}

		return fileCopyIdMaps, nil
	})
	if err != nil {
		log.Println("rfsCmdModel.go: Copy:", err)
		return nil, fiber.ErrInternalServerError
	}

	return res.([]any), nil
}
