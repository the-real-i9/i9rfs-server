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

			bd, err := resBody(res.Body)
			assert.NoError(t, err)
			assert.NotEmpty(t, bd)

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

	nativeRootDirs := make(map[string]string)

	t.Run("list native directories in root", func(t *testing.T) {

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

		dirMaps, ok := wsResp.Body.([]any)
		assert.True(t, ok)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			nativeRootDirs[m["name"].(string)] = m["id"].(string)
		}

		assert.Contains(t, nativeRootDirs, "Documents")
		assert.Contains(t, nativeRootDirs, "Downloads")
		assert.Contains(t, nativeRootDirs, "Videos")
		assert.Contains(t, nativeRootDirs, "Music")
		assert.Contains(t, nativeRootDirs, "Pictures")
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

			assert.NoError(t, wsConn.WriteJSON(sendData))

			var wsResp appTypes.WSResp

			assert.NoError(t, wsConn.ReadJSON(&wsResp))

			assert.Equal(t, http.StatusOK, wsResp.StatusCode)
			assert.NotEmpty(t, wsResp.Body)
			assert.Empty(t, wsResp.ErrorMsg)
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

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)

		dirMaps, ok := wsResp.Body.([]any)
		assert.True(t, ok)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			videoDirs[m["name"].(string)] = m["id"].(string)
		}
		assert.Contains(t, videoDirs, "Horror")
		assert.Contains(t, videoDirs, "Comedy")
		assert.Contains(t, videoDirs, "Legal")
		assert.Contains(t, videoDirs, "Musical")
		assert.Contains(t, videoDirs, "Action")
		assert.Contains(t, videoDirs, "NotAVideo")
		assert.Contains(t, videoDirs, "DeleteMe")
	})

	t.Run("delete 'NotAVideo' and 'DeleteMe' dirs in native root dir 'Videos'", func(t *testing.T) {
		sendData := map[string]any{
			"command": "del",
			"data": map[string]any{
				"parentDirectoryId": nativeRootDirs["Videos"],
				"objectIds":         []string{videoDirs["NotAVideo"], videoDirs["DeleteMe"]},
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)
	})

	t.Run("list the dirs now in native dir 'Videos' to confirm deletion", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": nativeRootDirs["Videos"],
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)

		dirMaps, ok := wsResp.Body.([]any)
		assert.True(t, ok)

		clear(videoDirs)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			videoDirs[m["name"].(string)] = m["id"].(string)
		}
		assert.Contains(t, videoDirs, "Musical")
		assert.Contains(t, videoDirs, "Action")
		assert.NotContains(t, videoDirs, "NotAVideo")
		assert.NotContains(t, videoDirs, "DeleteMe")
	})

	t.Run("attempt to delete a native directory fails", func(t *testing.T) {
		sendData := map[string]any{
			"command": "del",
			"data": map[string]any{
				"parentDirectoryId": "/",
				"objectIds":         []string{nativeRootDirs["Downloads"]},
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusBadRequest, wsResp.StatusCode)
		assert.Empty(t, wsResp.Body)
		assert.NotEmpty(t, wsResp.ErrorMsg)
	})

	assert.NoError(t, wsConn.CloseHandler()(websocket.CloseNormalClosure, "done"))

}

func TestCmds_CaseTwo(t *testing.T) {
	userSessionCookie := ""
	t.Run("user signup", func(t *testing.T) {
		signupSessCookie := ""

		t.Run("request new account", func(t *testing.T) {
			reqBody, err := reqBody(map[string]any{"email": "harveyspecter@gmail.com"})
			assert.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			bd, err := resBody(res.Body)
			assert.NoError(t, err)
			assert.NotEmpty(t, bd)

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
				"username": "harvey",
				"password": "scottie_",
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

	nativeRootDirs := make(map[string]string)

	t.Run("list native directories in root", func(t *testing.T) {

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

		dirMaps, ok := wsResp.Body.([]any)
		assert.True(t, ok)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			nativeRootDirs[m["name"].(string)] = m["id"].(string)
		}

		assert.Contains(t, nativeRootDirs, "Documents")
		assert.Contains(t, nativeRootDirs, "Downloads")
		assert.Contains(t, nativeRootDirs, "Videos")
		assert.Contains(t, nativeRootDirs, "Music")
		assert.Contains(t, nativeRootDirs, "Pictures")
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

			assert.NoError(t, wsConn.WriteJSON(sendData))

			var wsResp appTypes.WSResp

			assert.NoError(t, wsConn.ReadJSON(&wsResp))

			assert.Equal(t, http.StatusOK, wsResp.StatusCode)
			assert.NotEmpty(t, wsResp.Body)
			assert.Empty(t, wsResp.ErrorMsg)
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

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)

		dirMaps, ok := wsResp.Body.([]any)
		assert.True(t, ok)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			musicDirs[m["name"].(string)] = m["id"].(string)
		}
		assert.Contains(t, musicDirs, "Rock")
		assert.Contains(t, musicDirs, "Gospel")
		assert.Contains(t, musicDirs, "Pop")
		assert.Contains(t, musicDirs, "Folk")
		assert.Contains(t, musicDirs, "Old Songs")
	})

	t.Run("trash 'Folk' and 'Old Songs' dirs in 'Music'", func(t *testing.T) {

		sendData := map[string]any{
			"command": "trash",
			"data": map[string]any{
				"parentDirectoryId": nativeRootDirs["Music"],
				"objectIds":         []string{musicDirs["Folk"], musicDirs["Old Songs"]},
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)
	})

	t.Run("list the dirs now in 'Music': confirm trashing", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": nativeRootDirs["Music"],
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)

		dirMaps, ok := wsResp.Body.([]any)
		assert.True(t, ok)

		clear(musicDirs)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			musicDirs[m["name"].(string)] = m["id"].(string)
		}
		assert.Contains(t, musicDirs, "Pop")
		assert.Contains(t, musicDirs, "Gospel")
		assert.NotContains(t, musicDirs, "Folk")
		assert.NotContains(t, musicDirs, "Old Songs")
	})

	trashDirs := make(map[string]string)

	t.Run("view the dirs in Trash", func(t *testing.T) {

		sendData := map[string]any{
			"command": "view trash",
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)

		dirMaps, ok := wsResp.Body.([]any)
		assert.True(t, ok)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			trashDirs[m["name"].(string)] = m["id"].(string)
		}
		assert.Contains(t, trashDirs, "Folk")
		assert.Contains(t, trashDirs, "Old Songs")
	})

	t.Run("restore 'Folk' dir from Trash", func(t *testing.T) {

		sendData := map[string]any{
			"command": "restore",
			"data": map[string]any{
				"objectIds": []string{trashDirs["Folk"]},
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)
	})

	t.Run("list the dirs now in 'Music': confirm 'Folk' restored", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": nativeRootDirs["Music"],
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)

		dirMaps, ok := wsResp.Body.([]any)
		assert.True(t, ok)

		clear(musicDirs)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			musicDirs[m["name"].(string)] = m["id"].(string)
		}
		assert.Contains(t, musicDirs, "Pop")
		assert.Contains(t, musicDirs, "Gospel")
		assert.Contains(t, musicDirs, "Folk")
		assert.NotContains(t, musicDirs, "Old Songs")
	})

	t.Run("attempt to trash a native directory fails", func(t *testing.T) {
		sendData := map[string]any{
			"command": "trash",
			"data": map[string]any{
				"parentDirectoryId": "/",
				"objectIds":         []string{nativeRootDirs["Documents"]},
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusBadRequest, wsResp.StatusCode)
		assert.Empty(t, wsResp.Body)
		assert.NotEmpty(t, wsResp.ErrorMsg)
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

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)
	})

	t.Run("list the dirs now in 'Music': confirm renaming", func(t *testing.T) {

		sendData := map[string]any{
			"command": "ls",
			"data": map[string]any{
				"directoryId": nativeRootDirs["Music"],
			},
		}

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusOK, wsResp.StatusCode)
		assert.NotEmpty(t, wsResp.Body)
		assert.Empty(t, wsResp.ErrorMsg)

		dirMaps, ok := wsResp.Body.([]any)
		assert.True(t, ok)

		clear(musicDirs)

		for _, dm := range dirMaps {
			m := dm.(map[string]any)
			musicDirs[m["name"].(string)] = m["id"].(string)
		}
		assert.NotContains(t, musicDirs, "Gospel")
		assert.Contains(t, musicDirs, "Christian Music")
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

		assert.NoError(t, wsConn.WriteJSON(sendData))

		var wsResp appTypes.WSResp

		assert.NoError(t, wsConn.ReadJSON(&wsResp))

		assert.Equal(t, http.StatusBadRequest, wsResp.StatusCode)
		assert.Empty(t, wsResp.Body)
		assert.NotEmpty(t, wsResp.ErrorMsg)
	})

	assert.NoError(t, wsConn.CloseHandler()(websocket.CloseNormalClosure, "done"))
}
