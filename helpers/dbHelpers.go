package helpers

import (
	"context"
	"i9rfs/server/appGlobals"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func MultiOpQuery(transactionQueries func(ctx context.Context) (any, error)) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	txnOpts := options.Transaction().SetReadConcern(readconcern.Majority())

	opts := options.Session().SetDefaultTransactionOptions(txnOpts)

	sess, err := appGlobals.DB.Client().StartSession(opts)
	if err != nil {
		log.Panic(err)
	}
	defer sess.EndSession(ctx)

	txnOpts.SetReadPreference(readpref.PrimaryPreferred())

	result, err := sess.WithTransaction(ctx, transactionQueries, txnOpts)

	return result, err
}

func QueryRowField[T any](sql string, params ...any) (*T, error) {

	return nil, nil
}

func QueryRowsField[T any](sql string, params ...any) ([]*T, error) {

	return nil, nil
}

func QueryRowType[T any](sql string, params ...any) (*T, error) {

	return nil, nil
}

func QueryRowsType[T any](sql string, params ...any) ([]*T, error) {

	return nil, nil
}

func BatchQuery[T any](sqls []string, params [][]any) ([]*T, error) {

	return nil, nil
}
