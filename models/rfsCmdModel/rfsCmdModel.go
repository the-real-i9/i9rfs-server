package rfsCmdModel

import (
	"context"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func toGraphPath(objPath string) (string, map[string]any, string) {
	pathTree := strings.Split(objPath, "/")[1:]

	graphPath := ""
	paramMap := make(map[string]any)
	lastNodeIdent := ""

	for i, objName := range pathTree {
		if graphPath != "" {
			graphPath += "-[:HAS_CHILD]->"
		}

		nameKey := fmt.Sprintf("dir_%d", i)

		graphPath += fmt.Sprintf("(%s:Object{ name: $%[1]s })", nameKey)

		paramMap[nameKey] = objName

		lastNodeIdent = nameKey
	}

	return graphPath, paramMap, lastNodeIdent
}

func PathExists(ctx context.Context, path string) (bool, error) {
	graphPath, paramMap, _ := toGraphPath(path)

	res, err := neo4j.ExecuteQuery(ctx, appGlobals.Neo4jDriver,
		fmt.Sprintf(`
			RETURN EXISTS {
				MATCH %s
			} AS path_exists
		`, graphPath),
		paramMap,
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithReadersRouting(),
	)
	if err != nil {
		log.Println(fmt.Errorf("rfsCmdModel.go: PathExists: %s", err))
		return false, appGlobals.ErrInternalServerError
	}

	exists, _, err := neo4j.GetRecordValue[bool](res.Records[0], "path_exists")
	if err != nil {
		log.Println(fmt.Errorf("rfsCmdModel.go: PathExists: %s", err))
		return false, appGlobals.ErrInternalServerError
	}

	return exists, nil
}

type execResult struct {
	Status bool
	ErrMsg string
}

type cmdDBRes struct {
	Status bool
	ErrMsg string `db:"err_msg"`
}

func Mkdir(ctx context.Context, workPath string, newDirTree []string, clientUsername string) (bool, error) {
	sess := appGlobals.Neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})

	defer func() {
		if err := sess.Close(ctx); err != nil {
			log.Println("rfsCmdModel.go: Mkdir: sess.Close: error closing session")
		}
	}()

	res, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		graphPath, paramMap, lastNodeIdent := toGraphPath(workPath)

		paramMap["client_username"] = clientUsername

		res, err := tx.Run(ctx,
			fmt.Sprintf(`
				OPTIONAL MATCH %s
				WHERE obj_0.owner_user = $client_username
				RETURN %s.id AS parent_dir_id
				`, graphPath, lastNodeIdent),
			paramMap,
		)
		if err != nil {
			return nil, err
		}

		workDirId, _, _ := neo4j.GetRecordValue[string](res.Record(), "parent_dir_id")

		if workDirId == "" {
			return execResult{
				Status: false,
				ErrMsg: fmt.Sprintf("non-existent work path %s", workPath),
			}, nil
		}

		currDirId := workDirId

		for _, dirName := range newDirTree {
			res, err := tx.Run(ctx,
				`
				MATCH (currDir:Object:Directory{ id: $curr_dir_id })
				MERGE (currDir)->[:HAS_CHILD]->(childDir:Object:Directory{ name: $dir_name, object_type: "directory" }))
				ON CREATE
					SET childDir.id = randomUUID(), childDir.owner_user = $client_username, childDir.date_created = datetime(), childDir.date_modified = datetime()
				RETURN childDir.id AS child_dir_id
				`,
				map[string]any{"curr_dir_id": currDirId, "client_username": clientUsername, "dir_name": dirName},
			)
			if err != nil {
				return nil, err
			}

			currDirId, _, _ = neo4j.GetRecordValue[string](res.Record(), "child_dir_id")
		}

		return execResult{
			Status: true,
			ErrMsg: "",
		}, nil
	})

	if err != nil {
		log.Println("rfsCmdModel.go: Mkdir:", err)
		return false, appGlobals.ErrInternalServerError
	}

	execRes := res.(execResult)

	if !execRes.Status {
		return false, fmt.Errorf("%s", execRes.ErrMsg)
	}

	return true, nil
}

func Rmdir(ctx context.Context, dirPath, dirPathCmdArg, clientUsername string) (bool, error) {
	sess := appGlobals.Neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})

	defer func() {
		if err := sess.Close(ctx); err != nil {
			log.Println("rfsCmdModel.go: Mkdir: sess.Close: error closing session")
		}
	}()

	res, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		graphPath, paramMap, lastNodeIdent := toGraphPath(dirPath)

		paramMap["client_username"] = clientUsername

		res, err := tx.Run(ctx,
			fmt.Sprintf(`
				OPTIONAL MATCH %s
				WHERE obj_0.owner_user = $client_username

				WITH %[2]s, %[2]s IS NULL AS not_exist, %[2]s.object_type = "directory" AS is_directory, EXISTS { (%[2]s)-[:HAS_CHILD]->() } AS has_child

				DETACH DELETE %[2]s

				RETURN { not_exist, is_directory, has_child } AS dir_checks
				`, graphPath, lastNodeIdent),
			paramMap,
		)
		if err != nil {
			return nil, err
		}

		var dirChecks struct {
			NotExist    bool `json:"not_exist"`
			IsDirectory bool `json:"is_directory"`
			HasChild    bool `json:"has_child"`
		}

		dirChecksMap, _, _ := neo4j.GetRecordValue[map[string]any](res.Record(), "dir_checks")

		helpers.MapToStruct(dirChecksMap, &dirChecks)

		if dirChecks.NotExist {
			return execResult{
				Status: false,
				ErrMsg: fmt.Sprintf("failed to remove '%s': No such file or directory", dirPathCmdArg),
			}, nil
		}

		if !dirChecks.IsDirectory {
			return execResult{
				Status: false,
				ErrMsg: fmt.Sprintf("failed to remove '%s': Not a directory", dirPathCmdArg),
			}, nil
		}

		if dirChecks.HasChild {
			return execResult{
				Status: false,
				ErrMsg: fmt.Sprintf("failed to remove '%s': Directory not empty", dirPathCmdArg),
			}, nil
		}

		return execResult{
			Status: true,
			ErrMsg: "",
		}, nil
	})
	if err != nil {
		log.Println("rfsCmdModel.go: Rmdir:", err)
		return false, appGlobals.ErrInternalServerError
	}

	execRes := res.(execResult)

	if !execRes.Status {
		return false, fmt.Errorf("%s", execRes.ErrMsg)
	}

	return true, nil
}

func Rm(ctx context.Context, objectPath string, recursive bool, objectPathCmdArg, clientUsername string) (bool, []string, error) {
	sess := appGlobals.Neo4jDriver.NewSession(ctx, neo4j.SessionConfig{})

	defer func() {
		if err := sess.Close(ctx); err != nil {
			log.Println("rfsCmdModel.go: Rm: sess.Close: error closing session")
		}
	}()

	type execResult struct {
		Status  bool
		FileIds []string
		ErrMsg  string
	}

	res, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		graphPath, paramMap, lastNodeIdent := toGraphPath(objectPath)

		paramMap["client_username"] = clientUsername

		res, err := tx.Run(ctx,
			fmt.Sprintf(`
				OPTIONAL MATCH %s
				WHERE obj_0.owner_user = $client_username

				WITH %[2]s IS NULL AS not_exist, %[2]s.object_type = "directory" AS is_directory, %[2]s.id AS obj_id

				RETURN { not_exist, is_directory, obj_id } AS obj_res
				`, graphPath, lastNodeIdent),
			paramMap,
		)
		if err != nil {
			return nil, err
		}

		var objRes struct {
			NotExist    bool   `json:"not_exist"`
			IsDirectory bool   `json:"is_directory"`
			Id          string `json:"obj_id"`
		}

		objResMap, _, _ := neo4j.GetRecordValue[map[string]any](res.Record(), "obj_res")

		helpers.MapToStruct(objResMap, &objRes)

		if objRes.NotExist {
			return execResult{
				ErrMsg: fmt.Sprintf("cannot remove '%s': No such file or directory", objectPathCmdArg),
			}, nil
		}

		if objRes.IsDirectory && !recursive {
			return execResult{
				ErrMsg: fmt.Sprintf("cannot remove '%s': Is a directory", objectPathCmdArg),
			}, nil
		}

		// recursive remove: removes node, and all its children (if object_type = "directory")
		// returns all file (object_type) ids
		res2, err := tx.Run(ctx,
			`
				MATCH (obj:Object{ id: $obj_id })(()-[:HAS_CHILD]->(childObj))*
	
				WITH obj, childObj, CASE obj.object_type WHEN = "file" THEN [obj.id] ELSE [] END AS objFileId, [co IN childObj WHERE co.object_type = "file" | co.id] AS childObjFileIds

				DETACH DELETE obj, childObj
	
				RETURN objFileId + childObjFileIds AS file_ids
				`,
			map[string]any{"obj_id": objRes.Id},
		)
		if err != nil {
			return nil, err
		}

		fileIds, _ := res2.Record().Get("file_ids")

		return execResult{
			Status:  true,
			FileIds: fileIds.([]string),
		}, nil
	})
	if err != nil {
		log.Println("rfsCmdModel.go: Rm:", err)
		return false, nil, appGlobals.ErrInternalServerError
	}

	execRes := res.(execResult)

	if !execRes.Status {
		return false, nil, fmt.Errorf("%s", execRes.ErrMsg)
	}

	return true, execRes.FileIds, nil
}

func Mv(sourcePath, destPath string) (bool, error) {

	return true, nil
}
