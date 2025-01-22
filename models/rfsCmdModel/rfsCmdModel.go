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

func PathExists(ctx context.Context, path string) (bool, error) {
	pathTree := strings.Split(path, "/")[1:]

	graphPath := ""
	paramMap := make(map[string]any)

	for i, object_name := range pathTree {
		if graphPath != "" {
			graphPath += "-[:HAS_CHILD]->"
		}

		nameKey := fmt.Sprintf("obj_%d", i)

		graphPath += fmt.Sprintf("(:Object{ name: %s })", nameKey)

		paramMap[nameKey] = object_name
	}

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

func Mkdir(parentDirPath string, newDirTree []string, userId string) (bool, error) {
	res, err := helpers.QueryRowType[cmdDBRes]("SELECT status, err_msg FROM mkdir($1, $2, $3)", parentDirPath, newDirTree, userId)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		log.Println(fmt.Errorf("rfsCmdModel.go: Mkdir: %s: %s", pgErr.Message, pgErr.Detail))
		return false, appGlobals.ErrInternalServerError
	}

	if !res.Status {
		return false, fmt.Errorf("%s", res.ErrMsg)
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
