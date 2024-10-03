package initializers

import (
	"context"
	"i9rfs/server/appGlobals"
	"os"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background())
	if err != nil {
		return err
	}

	appGlobals.GCSClient = stClient
	return nil
}

func initDBClient() error {
	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("MONGODB_URL")))

	if err != nil {
		return err
	}

	appGlobals.DB = client.Database(os.Getenv("MONGODB_DB"))

	return nil
}

func InitApp() error {

	if err := godotenv.Load(".env"); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	if err := initDBClient(); err != nil {
		return err
	}

	return nil
}
