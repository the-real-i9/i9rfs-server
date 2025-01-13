package initializers

import (
	"context"
	"i9rfs/server/appGlobals"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background())
	if err != nil {
		return err
	}

	appGlobals.GCSClient = stClient
	return nil
}

/* func initDBPool() error {
	pool, err := pgxpool.New(context.Background(), os.Getenv("PGDATABASE_URL"))
	if err != nil {
		return err
	}
	appGlobals.DBPool = pool

	return nil
} */

func initNeo4jDriver() error {
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URL"), neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		return err
	}

	appGlobals.Neo4jDriver = driver

	return nil
}

func InitApp() error {

	if err := godotenv.Load(".env"); err != nil {
		return err
	}

	if err := initGCSClient(); err != nil {
		return err
	}

	/* if err := initDBPool(); err != nil {
		return err
	} */

	if err := initNeo4jDriver(); err != nil {
		return err
	}

	return nil
}

func CleanUp() {
	err := appGlobals.Neo4jDriver.Close(context.TODO())
	if err != nil {
		log.Println("error closing neo4j driver", err)
	}
}
