package tests

import (
	"i9rfs/appTypes"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/fasthttp/websocket"
	"github.com/stretchr/testify/require"
)

const WS_URL string = "ws://localhost:8000/api/app/rfs"

func TestCmds_CaseOne(t *testing.T) {
	userSessionCookie := ""
	t.Run("user signup", func(t *testing.T) {
		signupSessCookie := ""

		t.Run("request new account", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"email": "mikeross@gmail.com"})
			require.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			signupSessCookie = res.Header.Get("Set-Cookie")
		})

		t.Run("sends the correct email verf code", func(t *testing.T) {
			verfCode, err := strconv.Atoi(os.Getenv("DUMMY_VERF_TOKEN"))
			require.NoError(t, err)

			reqBody, err := reqBody(map[string]any{"code": verfCode})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
			require.NoError(t, err)
			req.Header.Set("Cookie", signupSessCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, res.StatusCode)
		})

		t.Run("submits her remaining credentials", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{
				"username": "mikeross",
				"password": "paralegal_zane",
			})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
			require.NoError(t, err)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Set("Cookie", signupSessCookie)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, res.StatusCode)

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
		require.NoError(t, err)
		require.Equal(t, http.StatusSwitchingProtocols, res.StatusCode)
	})

	nativeRootDirs := make(map[string]string)

	t.Run("exec command: ls: on root dir", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": "/",
			},
		}

		require.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		require.NoError(t, wsConn.ReadJSON(&wsResp))

		require.Equal(t, http.StatusOK, wsResp.StatusCode)
		require.NotEmpty(t, wsResp.Body)
		require.Empty(t, wsResp.ErrorMsg)

		dirMaps, ok := wsResp.Body.([]any)
		require.True(t, ok)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			nativeRootDirs[m["name"].(string)] = m["id"].(string)
		}

		require.Contains(t, nativeRootDirs, "Documents")
		require.Contains(t, nativeRootDirs, "Downloads")
		require.Contains(t, nativeRootDirs, "Videos")
		require.Contains(t, nativeRootDirs, "Music")
		require.Contains(t, nativeRootDirs, "Pictures")
	})

	t.Run("bulk create several dirs in native dirs", func(t *testing.T) {
		for pd, cds := range map[string][]string{"Videos": {"Horror", "Comedy", "Legal", "Musical", "Action", "NotAVideo", "DeleteMe"}, "Music": {"Gospel", "Rock", "Pop", "Folk", "Old Songs"}, "Pictures": {"Cats", "Photoshopped", "Landscape", "Girls", "Animes"}} {
			for _, cdir := range cds {
				sendData := map[string]any{
					"command": "mkdir",
					"data": map[string]any{
						"parentDirectoryId": nativeRootDirs[pd],
						"directoryName":     cdir,
					},
				}

				require.NoError(t, wsConn.WriteJSON(sendData))

				var wsResp appTypes.WSResp

				require.NoError(t, wsConn.ReadJSON(&wsResp))

				require.Equal(t, http.StatusOK, wsResp.StatusCode)
				require.NotEmpty(t, wsResp.Body)
				require.Empty(t, wsResp.ErrorMsg)
			}
		}
	})

	videoDirs := make(map[string]string)
	t.Run("list the dirs now in native dir 'Videos'", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": nativeRootDirs["Videos"],
			},
		}

		require.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		require.NoError(t, wsConn.ReadJSON(&wsResp))

		require.Equal(t, http.StatusOK, wsResp.StatusCode)
		require.NotEmpty(t, wsResp.Body)
		require.Empty(t, wsResp.ErrorMsg)

		dirMaps, ok := wsResp.Body.([]any)
		require.True(t, ok)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			videoDirs[m["name"].(string)] = m["id"].(string)
		}
		require.Contains(t, videoDirs, "Horror")
		require.Contains(t, videoDirs, "Comedy")
		require.Contains(t, videoDirs, "Legal")
		require.Contains(t, videoDirs, "Musical")
		require.Contains(t, videoDirs, "Action")
		require.Contains(t, videoDirs, "NotAVideo")
		require.Contains(t, videoDirs, "DeleteMe")
	})

	t.Run("delete 'NotAVideo' dir in native root dir 'Videos'", func(t *testing.T) {
		sendData := map[string]any{
			"command": "del",
			"data": map[string]any{
				"parentDirectoryId": nativeRootDirs["Videos"],
				"objectIds":         []string{videoDirs["NotAVideo"], videoDirs["DeleteMe"]},
			},
		}

		require.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		require.NoError(t, wsConn.ReadJSON(&wsResp))

		require.Equal(t, http.StatusOK, wsResp.StatusCode)
		require.NotEmpty(t, wsResp.Body)
		require.Empty(t, wsResp.ErrorMsg)
	})

	require.NoError(t, wsConn.CloseHandler()(websocket.CloseNormalClosure, "done"))

	// cleanUpDB()
}
