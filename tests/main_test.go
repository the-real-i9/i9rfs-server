// User-story-based testing for server applications
package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"testing"

	"github.com/fasthttp/websocket"
	"github.com/maxatome/go-testdeep/td"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const HOST_URL string = "http://localhost:8000"

const signupPath = HOST_URL + "/api/auth/signup"
const signinPath = HOST_URL + "/api/auth/signin"

const signoutPath = HOST_URL + "/api/app/signout"

const rfsPath string = "ws://localhost:8000/api/app/rfs"

type UserT struct {
	Email         string
	Username      string
	Password      string
	SessionCookie string
	WSConn        *websocket.Conn
	ServerWSMsg   chan map[string]any
}

func TestMain(m *testing.M) {
	dbDriver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URL"), neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	defer dbDriver.Close(ctx)

	neo4j.ExecuteQuery(ctx, dbDriver, `MATCH (n) DETACH DELETE n`, nil, neo4j.EagerResultTransformer)

	c := m.Run()

	os.Exit(c)
}

func makeReqBody(data map[string]any) (io.Reader, error) {
	dataBt, err := json.Marshal(data)

	return bytes.NewReader(dataBt), err
}

func succResBody[T any](body io.ReadCloser) (T, error) {
	var d T

	defer body.Close()

	bt, err := io.ReadAll(body)
	if err != nil {
		return d, err
	}

	if err := json.Unmarshal(bt, &d); err != nil {
		return d, err
	}

	return d, nil
}

func errResBody(body io.ReadCloser) (string, error) {
	defer body.Close()

	bt, err := io.ReadAll(body)
	if err != nil {
		return "", err
	}

	return string(bt), nil
}

func containsDirs(dir ...string) td.TestDeep {
	containsList := make([]any, len(dir))

	for i, dirName := range dir {
		containsList[i] = td.Contains(td.SuperMapOf(map[string]any{
			"id":       td.Ignore(),
			"obj_type": "directory",
			"name":     dirName,
		}, nil))
	}

	return td.All(containsList...)
}

func notContainsDirs(dir ...string) td.TestDeep {
	notContainsList := make([]any, len(dir))

	for i, dirName := range dir {
		notContainsList[i] = td.Not(td.Contains(td.SuperMapOf(map[string]any{
			"id":       td.Ignore(),
			"obj_type": "directory",
			"name":     dirName,
		}, nil)))
	}

	return td.All(notContainsList...)
}
