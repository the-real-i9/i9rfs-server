package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
)

const HOST_URL string = "http://localhost:8000"

var dbDriver neo4j.DriverWithContext

func TestMain(m *testing.M) {
	driver, err := neo4j.NewDriverWithContext(os.Getenv("NEO4J_URL"), neo4j.BasicAuth(os.Getenv("NEO4J_USER"), os.Getenv("NEO4J_PASSWORD"), ""))
	if err != nil {
		log.Fatalln(err)
	}

	dbDriver = driver

	ctx := context.Background()

	defer dbDriver.Close(ctx)

	c := m.Run()

	os.Exit(c)
}

func cleanUpDB() {
	neo4j.ExecuteQuery(context.Background(), dbDriver, `MATCH (n) DETACH DELETE n`, nil, neo4j.EagerResultTransformer)
}

func reqBody(data map[string]any) (io.Reader, error) {
	dataBt, err := json.Marshal(data)

	return bytes.NewReader(dataBt), err
}

func resBody(body io.ReadCloser) ([]byte, error) {
	defer body.Close()

	return io.ReadAll(body)
}

func TestSignup(t *testing.T) {
	signupPath := HOST_URL + "/api/auth/signup"

	cookie := ""

	t.Run("request new account", func(t *testing.T) {
		reqBody, err := reqBody(map[string]any{"email": "suberu@gmail.com"})
		assert.NoError(t, err)

		res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, res.StatusCode)

		cookie = res.Header.Get("Set-Cookie")

		resBody, err := resBody(res.Body)
		assert.NoError(t, err)
		t.Logf("%s", resBody)
	})

	t.Run("verify email", func(t *testing.T) {
		verfCode, err := strconv.Atoi(os.Getenv("DUMMY_VERF_TOKEN"))
		assert.NoError(t, err)

		reqBody, err := reqBody(map[string]any{"code": verfCode})
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
		assert.NoError(t, err)
		req.Header.Set("Cookie", cookie)
		req.Header.Add("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		resBody, err := resBody(res.Body)
		assert.NoError(t, err)
		t.Logf("%s", resBody)
	})

	t.Run("register user", func(t *testing.T) {
		reqBody, err := reqBody(map[string]any{
			"username": "suberu",
			"password": "sketeppy",
		})
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
		assert.NoError(t, err)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Set("Cookie", cookie)

		res, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)

		resBody, err := resBody(res.Body)
		assert.NoError(t, err)
		t.Logf("%s", resBody)
	})

	cleanUpDB()
}
