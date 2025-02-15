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
	t.Parallel()

	userSessionCookie := ""
	t.Run("user signup", func(t *testing.T) {
		signupSessCookie := ""

		t.Run("request new account", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"email": "mikeross@gmail.com"})
			require.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			bd, err := resBody(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, bd)

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

	t.Run("list native directories in root", func(t *testing.T) {

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

	t.Run("bulk create dirs in 'Videos' dir", func(t *testing.T) {
		for _, dir := range []string{"Horror", "Comedy", "Legal", "Musical", "Action", "NotAVideo", "DeleteMe"} {
			sendData := map[string]any{
				"command": "mkdir",
				"data": map[string]any{
					"parentDirectoryId": nativeRootDirs["Videos"],
					"directoryName":     dir,
				},
			}

			require.NoError(t, wsConn.WriteJSON(sendData))

			var wsResp appTypes.WSResp

			require.NoError(t, wsConn.ReadJSON(&wsResp))

			require.Equal(t, http.StatusOK, wsResp.StatusCode)
			require.NotEmpty(t, wsResp.Body)
			require.Empty(t, wsResp.ErrorMsg)
		}
	})

	videoDirs := make(map[string]string)
	t.Run("list the dirs in native dir 'Videos'", func(t *testing.T) {

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

	t.Run("delete 'NotAVideo' and 'DeleteMe' dirs in native root dir 'Videos'", func(t *testing.T) {
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

	t.Run("list the dirs now in native dir 'Videos' to confirm deletion", func(t *testing.T) {

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

		clear(videoDirs)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			videoDirs[m["name"].(string)] = m["id"].(string)
		}
		require.Contains(t, videoDirs, "Musical")
		require.Contains(t, videoDirs, "Action")
		require.NotContains(t, videoDirs, "NotAVideo")
		require.NotContains(t, videoDirs, "DeleteMe")
	})

	t.Run("attempt to delete a native directory fails", func(t *testing.T) {
		sendData := map[string]any{
			"command": "del",
			"data": map[string]any{
				"parentDirectoryId": "/",
				"objectIds":         []string{nativeRootDirs["Downloads"]},
			},
		}

		require.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		require.NoError(t, wsConn.ReadJSON(&wsResp))

		require.Equal(t, http.StatusBadRequest, wsResp.StatusCode)
		require.Empty(t, wsResp.Body)
		require.NotEmpty(t, wsResp.ErrorMsg)
	})

	require.NoError(t, wsConn.CloseHandler()(websocket.CloseNormalClosure, "done"))

}

func TestCmds_CaseTwo(t *testing.T) {
	t.Parallel()

	userSessionCookie := ""
	t.Run("user signup", func(t *testing.T) {
		signupSessCookie := ""

		t.Run("request new account", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"email": "harveyspecter@gmail.com"})
			require.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)

			bd, err := resBody(res.Body)
			require.NoError(t, err)
			require.NotEmpty(t, bd)

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
				"username": "harvey",
				"password": "scottie_",
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

	t.Run("list native directories in root", func(t *testing.T) {

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

	t.Run("bulk create dirs in 'Music'", func(t *testing.T) {
		for _, cdir := range []string{"Gospel", "Rock", "Pop", "Folk", "Old Songs"} {
			sendData := map[string]any{
				"command": "mkdir",
				"data": map[string]any{
					"parentDirectoryId": nativeRootDirs["Music"],
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
	})

	musicDirs := make(map[string]string)
	t.Run("list the dirs in 'Music'", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": nativeRootDirs["Music"],
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
			musicDirs[m["name"].(string)] = m["id"].(string)
		}
		require.Contains(t, musicDirs, "Rock")
		require.Contains(t, musicDirs, "Gospel")
		require.Contains(t, musicDirs, "Pop")
		require.Contains(t, musicDirs, "Folk")
		require.Contains(t, musicDirs, "Old Songs")
	})

	t.Run("trash 'Folk' and 'Old Songs' dirs in 'Music'", func(t *testing.T) {

		sendData := map[string]any{
			"command": "trash",
			"data": map[string]any{
				"parentDirectoryId": nativeRootDirs["Music"],
				"objectIds":         []string{musicDirs["Folk"], musicDirs["Old Songs"]},
			},
		}

		require.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		require.NoError(t, wsConn.ReadJSON(&wsResp))

		require.Equal(t, http.StatusOK, wsResp.StatusCode)
		require.NotEmpty(t, wsResp.Body)
		require.Empty(t, wsResp.ErrorMsg)
	})

	t.Run("list the dirs now in 'Music': confirm trashing", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": nativeRootDirs["Music"],
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

		clear(musicDirs)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			musicDirs[m["name"].(string)] = m["id"].(string)
		}
		require.Contains(t, musicDirs, "Pop")
		require.Contains(t, musicDirs, "Gospel")
		require.NotContains(t, musicDirs, "Folk")
		require.NotContains(t, musicDirs, "Old Songs")
	})

	trashDirs := make(map[string]string)

	t.Run("view the dirs in Trash", func(t *testing.T) {

		sendData := map[string]any{
			"command": "view trash",
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
			trashDirs[m["name"].(string)] = m["id"].(string)
		}
		require.Contains(t, trashDirs, "Folk")
		require.Contains(t, trashDirs, "Old Songs")
	})

	t.Run("restore 'Folk' dir from Trash", func(t *testing.T) {

		sendData := map[string]any{
			"command": "restore",
			"data": map[string]any{
				"objectIds": []string{trashDirs["Folk"]},
			},
		}

		require.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		require.NoError(t, wsConn.ReadJSON(&wsResp))

		require.Equal(t, http.StatusOK, wsResp.StatusCode)
		require.NotEmpty(t, wsResp.Body)
		require.Empty(t, wsResp.ErrorMsg)
	})

	t.Run("list the dirs now in 'Music': confirm 'Folk' restored", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": nativeRootDirs["Music"],
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

		clear(musicDirs)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			musicDirs[m["name"].(string)] = m["id"].(string)
		}
		require.Contains(t, musicDirs, "Pop")
		require.Contains(t, musicDirs, "Gospel")
		require.Contains(t, musicDirs, "Folk")
		require.NotContains(t, musicDirs, "Old Songs")
	})

	t.Run("attempt to trash a native directory fails", func(t *testing.T) {
		sendData := map[string]any{
			"command": "trash",
			"data": map[string]any{
				"parentDirectoryId": "/",
				"objectIds":         []string{nativeRootDirs["Documents"]},
			},
		}

		require.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		require.NoError(t, wsConn.ReadJSON(&wsResp))

		require.Equal(t, http.StatusBadRequest, wsResp.StatusCode)
		require.Empty(t, wsResp.Body)
		require.NotEmpty(t, wsResp.ErrorMsg)
	})

	t.Run("rename 'Gospel' in 'Music': to 'Christian Music'", func(t *testing.T) {

		sendData := map[string]any{
			"command": "rename",
			"data": map[string]any{
				"parentDirectoryId": nativeRootDirs["Music"],
				"objectId":          musicDirs["Gospel"],
				"newName":           "Christian Music",
			},
		}

		require.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		require.NoError(t, wsConn.ReadJSON(&wsResp))

		require.Equal(t, http.StatusOK, wsResp.StatusCode)
		require.NotEmpty(t, wsResp.Body)
		require.Empty(t, wsResp.ErrorMsg)
	})

	t.Run("list the dirs now in 'Music': confirm renaming", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": nativeRootDirs["Music"],
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

		clear(musicDirs)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			musicDirs[m["name"].(string)] = m["id"].(string)
		}
		require.NotContains(t, musicDirs, "Gospel")
		require.Contains(t, musicDirs, "Christian Music")
	})

	t.Run("attempt to rename a native directory fails", func(t *testing.T) {
		sendData := map[string]any{
			"command": "rename",
			"data": map[string]any{
				"parentDirectoryId": "/",
				"objectId":          nativeRootDirs["Pictures"],
				"newName":           "Images",
			},
		}

		require.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		require.NoError(t, wsConn.ReadJSON(&wsResp))

		require.Equal(t, http.StatusBadRequest, wsResp.StatusCode)
		require.Empty(t, wsResp.Body)
		require.NotEmpty(t, wsResp.ErrorMsg)
	})

	require.NoError(t, wsConn.CloseHandler()(websocket.CloseNormalClosure, "done"))
}
