package appGlobals

import (
	"errors"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var ErrInternalServerError = errors.New("internal server error: check logger")

var GCSClient *storage.Client

var DBPool *pgxpool.Pool

var SignupSessionStore *session.Store

var Neo4jDriver neo4j.DriverWithContext
