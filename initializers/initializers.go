package initializers

import (
	"context"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/services/rfsCmdService"
	"os"
	"os/exec"

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
	appGlobals.DBClient = client

	return nil
}

func initAppDataStore() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	appHomeDir := fmt.Sprintf("%s/.i9rfs-server/home", homeDir)

	exec.Command("mkdir", "-p", appHomeDir).Run()

	rfsCmdService.SetHome(appHomeDir)
}

func InitApp() error {

	if err := godotenv.Load(".env"); err != nil {
		return err
	}

	if err := initDBClient(); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	initAppDataStore()

	return nil
}
