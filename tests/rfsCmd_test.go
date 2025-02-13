package tests

import (
	"i9rfs/appTypes"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/fasthttp/websocket"
	"github.com/stretchr/testify/assert"
)

const WS_URL string = "ws://localhost:8000/api/app/rfs"

func TestCmds_CaseOne(t *testing.T) {
	userSessionCookie := ""
	t.Run("user signup", func(t *testing.T) {
		signupSessCookie := ""

		t.Run("request new account", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"email": "mikeross@gmail.com"})
			assert.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			signupSessCookie = res.Header.Get("Set-Cookie")
		})

		t.Run("sends the correct email verf code", func(t *testing.T) {
			verfCode, err := strconv.Atoi(os.Getenv("DUMMY_VERF_TOKEN"))
			assert.NoError(t, err)

			reqBody, err := reqBody(map[string]any{"code": verfCode})
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
			assert.NoError(t, err)
			req.Header.Set("Cookie", signupSessCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)
		})

		t.Run("submits her remaining credentials", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{
				"username": "mikeross",
				"password": "paralegal_zane",
			})
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
			assert.NoError(t, err)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Cookie", signupSessCookie)

			res, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)

			userSessionCookie = res.Header.Get("Set-Cookie")
		})
	})

	var (
		wsConn *websocket.Conn
		res    *http.Response
		err    error
	)

	t.Run("websocket connect", func(t *testing.T) {
		wsHeader := http.Header{}
		wsHeader.Set("Cookie", userSessionCookie)
		wsConn, res, err = websocket.DefaultDialer.Dial(WS_URL, wsHeader)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusSwitchingProtocols, res.StatusCode)
	})

	sampleParentDirId := ""

	t.Run("exec command: ls", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": "/",
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)

		sampleParentDirId = wsResp.Body.([]any)[0].(map[string]any)["id"].(string)
		assert.NotEmpty(t, sampleParentDirId)
	})

	t.Run("exec command: mkdir", func(t *testing.T) {

		sendData := map[string]any{
			"command": "mkdir",
			"data": map[string]any{
				"parentDirectoryId": sampleParentDirId,
				"directoryName":     "folderA",
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)
	})

	t.Run("for sample parent directory above: exec command: ls", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": sampleParentDirId,
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)

		dirItemName := wsResp.Body.([]any)[0].(map[string]any)["name"].(string)
		assert.Contains(t, []string{"folderA"}, dirItemName)
	})

	assert.NoError(t, wsConn.CloseHandler()(websocket.CloseNormalClosure, "done"))

	cleanUpDB()
}
