package appGlobals

import (
	"errors"

	"cloud.google.com/go/storage"
	"github.com/gofiber/fiber/v2/middleware/session"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrInternalServerError = errors.New("internal server error: check logger")

var GCSClient *storage.Client

var DB *mongo.Database

var SignupSessionStore *session.Store
