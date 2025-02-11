package main

import (
	"context"
	"i9rfs/appGlobals"
	"testing"

	"github.com/fasthttp/websocket"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
)

func cleanDB() {
	neo4j.ExecuteQuery(context.Background(), appGlobals.Neo4jDriver,
		`
			MATCH (n)
			DETACH DELETE n
			`,
		nil,
		neo4j.EagerResultTransformer,
	)
}

func TestSignup(t *testing.T) {

	var sessionToken string

	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(context.Background(), "ws://localhost:8000", nil)
	assert.NoError(t, err)

	t.Run("request new account", func(t *testing.T) {

		w_err := conn.WriteJSON(map[string]any{
			"step": "one",
			"data": map[string]any{
				"email": "harvey@gmail.com",
			},
		})

		assert.NoError(t, w_err)

		r_err := conn.ReadJSON(&sessionToken)
		assert.NoError(t, r_err)

		assert.NotEmpty(t, sessionToken)
	})

	assert.NoError(t, conn.Close())
	cleanDB()
}
