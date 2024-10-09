package helpers

import (
	"context"
	"errors"
	"i9rfs/server/appGlobals"
	"time"

	"github.com/jackc/pgx/v5"
)

func QueryRowField[T any](sql string, params ...any) (*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, _ := appGlobals.DBPool.Query(ctx, sql, params...)

	res, err := pgx.CollectOneRow(rows, pgx.RowToAddrOf[T])
	if err != nil {
		return nil, err
	}

	return res, err
}

func QueryRowsField[T any](sql string, params ...any) ([]*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, _ := appGlobals.DBPool.Query(ctx, sql, params...)

	res, err := pgx.CollectRows(rows, pgx.RowToAddrOf[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return make([]*T, 0), nil
		}
		return nil, err
	}

	return res, nil
}

func QueryRowType[T any](sql string, params ...any) (*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, _ := appGlobals.DBPool.Query(ctx, sql, params...)

	res, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return res, nil
}

func QueryRowsType[T any](sql string, params ...any) ([]*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, _ := appGlobals.DBPool.Query(ctx, sql, params...)

	res, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByNameLax[T])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return res, nil
}

func BatchQuery[T any](sqls []string, params [][]any) ([]*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	var res = make([]*T, len(sqls))

	batch := &pgx.Batch{}

	for i, sql := range sqls {
		batch.Queue(sql, params[i]...).QueryRow(func(row pgx.Row) error {
			var sr *T

			if err := row.Scan(sr); err != nil {
				return err
			}

			res[i] = sr

			return nil
		})
	}

	s_err := appGlobals.DBPool.SendBatch(ctx, batch).Close()

	return res, s_err
}
