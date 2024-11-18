package rfsCmdModel

import (
	"errors"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/helpers"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
)

func PathExists(path string) (bool, error) {
	exists, err := helpers.QueryRowField[bool]("SELECT EXISTS(SELECT 1 FROM fs_object WHERE path = $1)", path)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		log.Println(fmt.Errorf("rfsCmdModel.go: PathExists: %s: %s", pgErr.Message, pgErr.Detail))
		return false, appGlobals.ErrInternalServerError
	}

	return *exists, nil
}

func Mkdir(parentDirPath string, newDirTree []string, userId string) (bool, error) {
	_, err := helpers.QueryRowField[bool]("SELECT mkdir($1, $2, $3)", parentDirPath, newDirTree, userId)
	if err != nil {
		var pgErr *pgconn.PgError
		errors.As(err, &pgErr)
		log.Println(fmt.Errorf("rfsCmdModel.go: Mkdir: %s: %s", pgErr.Message, pgErr.Detail))
		return false, appGlobals.ErrInternalServerError
	}

	return true, nil
}

type cmdDBRes struct {
	Status bool
	ErrMsg string `db:"err_msg"`
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
		return false, fmt.Errorf(res.ErrMsg)
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
		return false, fmt.Errorf(res.ErrMsg)
	}

	return true, nil
}
