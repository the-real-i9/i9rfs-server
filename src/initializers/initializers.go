package initializers

import (
	"context"
	"i9rfs/src/appGlobals"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"google.golang.org/api/option"
)

func initGCSClient() error {
	stClient, err := storage.NewClient(context.Background(), option.WithAPIKey(os.Getenv("GCS_API_KEY")))
	if err != nil {
		return err
	}

	appGlobals.GCSClient = stClient
	return nil
}

func initNeo4jDriver() error {
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URL"), neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sess := driver.NewSession(ctx, neo4j.SessionConfig{})

	_, err2 := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `CREATE CONSTRAINT unique_username IF NOT EXISTS FOR (u:User) REQUIRE u.username IS UNIQUE`, nil)
		if err != nil {
			return nil, err
		}

		_, err2 := tx.Run(ctx, `CREATE CONSTRAINT unique_email IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE`, nil)
		if err2 != nil {
			return nil, err2
		}

		_, err3 := tx.Run(ctx, `CREATE CONSTRAINT unique_object IF NOT EXISTS FOR (o:Object) REQUIRE o.id IS UNIQUE`, nil)
		if err3 != nil {
			return nil, err3
		}

		_, err4 := tx.Run(ctx, `CREATE CONSTRAINT unique_object_copy IF NOT EXISTS FOR (oc:Object) REQUIRE oc.copied_id IS UNIQUE`, nil)
		if err4 != nil {
			return nil, err4
		}

		return nil, nil
	})

	if err2 != nil {
		return err2
	}

	if err := sess.Close(ctx); err != nil {
		return err
	}

	appGlobals.Neo4jDriver = driver

	return nil
}

func InitApp() error {

	if os.Getenv("GO_ENV") == "" {
		if err := godotenv.Load(".env"); err != nil {
			return err
		}
	}

	if err := initGCSClient(); err != nil {
		return err
	}

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
