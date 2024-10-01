package helpers

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readconcern"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func MultiOpQuery(client *mongo.Client, transactionQueries func(ctx context.Context) (any, error)) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	txnOpts := options.Transaction().SetReadConcern(readconcern.Majority()) // transaction options

	opts := options.Session().SetDefaultTransactionOptions(txnOpts)

	// transaction options for session
	sess, err := client.StartSession(opts)
	if err != nil {
		log.Panic(err)
	}
	defer sess.EndSession(ctx)

	// extending transaction options with read preference (just before starting the transaction)
	// even if this is not set, "primary" will be the default
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
