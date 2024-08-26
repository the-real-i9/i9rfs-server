package initializers

import (
	"context"
	"fmt"
	"i9rfs/server/appGlobals"
	"i9rfs/server/services/rfsCmdService"
	"os"
	"os/exec"

	"cloud.google.com/go/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background(), option.WithCredentialsFile("i9apps-storage.json"))
	if err != nil {
		return err
	}

	appGlobals.GCSClient = stClient

	return nil
}

func initDBPool() error {
	pool, err := pgxpool.New(context.Background(), os.Getenv("PGDATABASE_URL"))
	if err != nil {
		return err
	}
	appGlobals.DBPool = pool

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

	if err := initDBPool(); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	initAppDataStore()

	return nil
}
