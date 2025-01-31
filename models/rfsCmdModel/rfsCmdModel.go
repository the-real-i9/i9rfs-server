package rfsCmdModel

import (
	"context"
	"errors"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func toGraphPath(dirPath string) (string, map[string]any, string) {
	pathTree := strings.Split(dirPath, "/")[1:]

	graphPath := ""
	paramMap := make(map[string]any)
	lastDirIdent := ""

	for i, dirName := range pathTree {
		if graphPath != "" {
			graphPath += "-[:HAS_CHILD]->"
		}

		nameKey := fmt.Sprintf("dir_%d", i)

		graphPath += fmt.Sprintf("(%s:Directory{ name: $%s })", nameKey)

		paramMap[nameKey] = dirName

		lastDirIdent = nameKey
	}

	return graphPath, paramMap, lastDirIdent
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

	type execResult struct {
		Status bool
		ErrMsg string
	}

	res, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		graphPath, paramMap, lastNodeIdent := toGraphPath(workPath)

		paramMap["client_username"] = clientUsername

		res1, err := tx.Run(ctx,
			fmt.Sprintf(`
				OPTIONAL MATCH %s
				WHERE dir_0.owner_user = $client_username
				RETURN %s.id AS parent_dir_id
				`, graphPath, lastNodeIdent),
			paramMap,
		)
		if err != nil {
			return nil, err
		}

		workDirId, _, _ := neo4j.GetRecordValue[string](res1.Record(), "parent_dir_id")

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
				MERGE (currDir)->[:HAS_CHILD]->(childDir:Object:Directory{ owner_user: $client_username, name: $dir_name }))
				ON CREATE
					SET childDir.id = randomUUID(), childDir.date_created = datetime(), childDir.date_modified = datetime()
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

func Rmdir(dirPath string) (bool, error) {
	res, err := helpers.QueryRowType[cmdDBRes]("SELECT status, err_msg FROM rmdir($1)", dirPath)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		log.Println(fmt.Errorf("rfsCmdModel.go: Rmdir: %s: %s", pgErr.Message, pgErr.Detail))
		return false, appGlobals.ErrInternalServerError
	}

	if !res.Status {
		return false, fmt.Errorf("%s", res.ErrMsg)
	}

	return true, nil
}

func Rm(fsObjectPath string, recursive bool) (bool, error) {
	res, err := helpers.QueryRowType[cmdDBRes]("SELECT status, err_msg FROM rm($1, $2)", fsObjectPath, recursive)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		log.Println(fmt.Errorf("rfsCmdModel.go: Rm: %s: %s", pgErr.Message, pgErr.Detail))
		return false, appGlobals.ErrInternalServerError
	}

	if !res.Status {
		return false, fmt.Errorf("%s", res.ErrMsg)
	}

	return true, nil
}

func Mv(sourcePath, destPath string) (bool, error) {
	res, err := helpers.QueryRowType[cmdDBRes]("SELECT status, err_msg FROM mv($1, $2)", sourcePath, destPath)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		log.Println(fmt.Errorf("rfsCmdModel.go: Mv: %s: %s", pgErr.Message, pgErr.Detail))
		return false, appGlobals.ErrInternalServerError
	}

	if !res.Status {
		return false, fmt.Errorf("%s", res.ErrMsg)
	}

	return true, nil
}
