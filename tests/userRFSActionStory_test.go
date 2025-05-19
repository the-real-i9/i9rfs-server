package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/fasthttp/websocket"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRFSActionStory(t *testing.T) {
	t.Parallel()

	user := UserT{
		Email:    "mikeross@gmail.com",
		Username: "mikeross",
		Password: "paralegal_zane",
	}

	{
		t.Log("Setup: create new account for users")

		{
			reqBody, err := makeReqBody(map[string]any{"email": user.Email})
			require.NoError(t, err)

			res, err := http.Post(signupPath+"/request_new_account", "application/json", reqBody)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(t, err)

			td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
				"msg": fmt.Sprintf("Enter the 6-digit code sent to %s to verify your email", user.Email),
			}, nil))

			user.SessionCookie = res.Header.Get("Set-Cookie")
		}

		{
			verfCode := os.Getenv("DUMMY_VERF_TOKEN")

			reqBody, err := makeReqBody(map[string]any{"code": verfCode})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/verify_email", reqBody)
			require.NoError(t, err)
			req.Header.Set("Cookie", user.SessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusOK, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(t, err)

			td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
				"msg": fmt.Sprintf("Your email, %s, has been verified!", user.Email),
			}, nil))

			user.SessionCookie = res.Header.Get("Set-Cookie")
		}

		{
			reqBody, err := makeReqBody(map[string]any{
				"username": user.Username,
				"password": user.Password,
			})
			require.NoError(t, err)

			req, err := http.NewRequest("POST", signupPath+"/register_user", reqBody)
			require.NoError(t, err)
			req.Header.Set("Cookie", user.SessionCookie)
			req.Header.Add("Content-Type", "application/json")

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			if !assert.Equal(t, http.StatusCreated, res.StatusCode) {
				rb, err := errResBody(res.Body)
				require.NoError(t, err)
				t.Log("unexpected error:", rb)
				return
			}

			rb, err := succResBody[map[string]any](res.Body)
			require.NoError(t, err)

			td.Cmp(td.Require(t), rb, td.SuperMapOf(map[string]any{
				"user": td.Ignore(),
				"msg":  "Signup success!",
			}, nil))

			user.SessionCookie = res.Header.Get("Set-Cookie")
		}
	}

	{
		t.Log("Setup: Init user sockets")

		header := http.Header{}
		header.Set("Cookie", user.SessionCookie)
		wsConn, res, err := websocket.DefaultDialer.Dial(rfsPath, header)
		require.NoError(t, err)

		if !assert.Equal(t, http.StatusSwitchingProtocols, res.StatusCode) {
			rb, err := errResBody(res.Body)
			require.NoError(t, err)
			t.Log("unexpected error:", rb)
			return
		}

		require.NotNil(t, wsConn)

		defer wsConn.CloseHandler()(websocket.CloseNormalClosure, user.Username+": GoodBye!")

		user.WSConn = wsConn
		user.ServerWSMsg = make(chan map[string]any)

		go func() {
			userCommChan := user.ServerWSMsg

			for {
				userCommChan := userCommChan
				userWSConn := user.WSConn

				var wsMsg map[string]any

				if err := userWSConn.ReadJSON(&wsMsg); err != nil {
					break
				}

				if wsMsg == nil {
					continue
				}

				userCommChan <- wsMsg
			}

			close(userCommChan)
		}()
	}

	t.Log("----------")

	nativeRootDirs := make(map[string]string, 5)

	{
		t.Log("Action: list native directories in root")

		err := user.WSConn.WriteJSON(map[string]any{
			"action": "ls",
			"data": map[string]any{
				"directoryId": "/",
			},
		})

		require.NoError(t, err)

		serverWSReply := <-user.ServerWSMsg

		td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
			"event":    "server reply",
			"toAction": "ls",
			"data":     containsDirs("Documents", "Downloads", "Music", "Videos", "Pictures"),
		}, nil))

		for _, dm := range serverWSReply["data"].([]any) {
			m := dm.(map[string]any)
			nativeRootDirs[m["name"].(string)] = m["id"].(string)
		}
	}

	{
		{
			t.Log("Action: bulk create dirs in native dir: 'Videos'")

			for _, dir := range []string{"Horror", "Comedy", "Legal", "Musical", "Action", "NotAVideo", "DeleteMe"} {
				err := user.WSConn.WriteJSON(map[string]any{
					"action": "mkdir",
					"data": map[string]any{
						"parentDirectoryId": nativeRootDirs["Videos"],
						"directoryName":     dir,
					},
				})

				require.NoError(t, err)

				serverWSReply := <-user.ServerWSMsg

				td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
					"event":    "server reply",
					"toAction": "mkdir",
					"data": td.SuperMapOf(map[string]any{
						"id":       td.Ignore(),
						"obj_type": "directory",
						"name":     dir,
					}, nil),
				}, nil))
			}
		}

		videoDirs := make(map[string]string, 7)

		{
			t.Log("list the dirs in native dir: 'Videos'")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "ls",
				"data": map[string]any{
					"directoryId": nativeRootDirs["Videos"],
				},
			})

			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "ls",
				"data":     containsDirs("Horror", "Comedy", "Legal", "Musical", "Action", "NotAVideo", "DeleteMe"),
			}, nil))

			for _, dm := range serverWSReply["data"].([]any) {
				m := dm.(map[string]any)
				videoDirs[m["name"].(string)] = m["id"].(string)
			}
		}

		{
			t.Log("Action: put a sub-directory inside 'DeleteMe' dir | to test recursive deletion")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "mkdir",
				"data": map[string]any{
					"parentDirectoryId": videoDirs["DeleteMe"],
					"directoryName":     "DeleteMyChild",
				},
			})

			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "mkdir",
				"data": td.SuperMapOf(map[string]any{
					"id":       td.Ignore(),
					"obj_type": "directory",
					"name":     "DeleteMyChild",
				}, nil),
			}, nil))
		}

		{
			t.Log("Action: delete 'NotAVideo' and 'DeleteMe' dirs in native root dir: 'Videos'")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "del",
				"data": map[string]any{
					"parentDirectoryId": nativeRootDirs["Videos"],
					"objectIds":         []string{videoDirs["NotAVideo"], videoDirs["DeleteMe"]},
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "del",
				"data":     true,
			}, nil))
		}

		{
			t.Log("list the dirs now in native dir: 'Videos' | confirm deletion")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "ls",
				"data": map[string]any{
					"directoryId": nativeRootDirs["Videos"],
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "ls",
				"data":     td.All(containsDirs("Horror", "Comedy", "Legal", "Musical", "Action"), notContainsDirs("NotAVideo", "DeleteMe")),
			}, nil))

			clear(videoDirs)

			for _, dm := range serverWSReply["data"].([]any) {
				m := dm.(map[string]any)
				videoDirs[m["name"].(string)] = m["id"].(string)
			}
		}

		{
			t.Log("attempt to delete a native directory fails")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "del",
				"data": map[string]any{
					"parentDirectoryId": "/",
					"objectIds":         []string{nativeRootDirs["Downloads"]},
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "del",
				"data":     false,
			}, nil))
		}

		{
			t.Log("Action: put sub-directories inside 'Horror' dir | to test recursive copy")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "mkdir",
				"data": map[string]any{
					"parentDirectoryId": videoDirs["Horror"],
					"directoryName":     "The Conjuring",
				},
			})

			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "mkdir",
				"data": td.SuperMapOf(map[string]any{
					"id":       td.Ignore(),
					"obj_type": "directory",
					"name":     "The Conjuring",
				}, nil),
			}, nil))

			err = user.WSConn.WriteJSON(map[string]any{
				"action": "mkdir",
				"data": map[string]any{
					"parentDirectoryId": serverWSReply["data"].(map[string]any)["id"],
					"directoryName":     "Season 1",
				},
			})

			require.NoError(t, err)

			serverWSReply = <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "mkdir",
				"data": td.SuperMapOf(map[string]any{
					"id":       td.Ignore(),
					"obj_type": "directory",
					"name":     "Season 1",
				}, nil),
			}, nil))

			err = user.WSConn.WriteJSON(map[string]any{
				"action": "mkdir",
				"data": map[string]any{
					"parentDirectoryId": serverWSReply["data"].(map[string]any)["id"],
					"directoryName":     "Episodes",
				},
			})

			require.NoError(t, err)

			serverWSReply = <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "mkdir",
				"data": td.SuperMapOf(map[string]any{
					"id":       td.Ignore(),
					"obj_type": "directory",
					"name":     "Episodes",
				}, nil),
			}, nil))
		}

		{
			t.Log("Action: copy 'Horror' and 'Comedy' dirs from/to native root dir 'Videos'/'Downloads'")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "copy",
				"data": map[string]any{
					"fromParentDirectoryId": nativeRootDirs["Videos"],
					"toParentDirectoryId":   nativeRootDirs["Downloads"],
					"objectIds":             []string{videoDirs["Horror"], videoDirs["Comedy"]},
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			t.Log(serverWSReply)

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "copy",
				"data":     true,
			}, nil))
		}

		{
			t.Log("list the dirs in native dir: 'Downloads' | confirm copy")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "ls",
				"data": map[string]any{
					"directoryId": nativeRootDirs["Downloads"],
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "ls",
				"data":     containsDirs("Horror", "Comedy"),
			}, nil))
		}
	}

	{
		{
			t.Log("Action: bulk create dirs in native dir: 'Music'")

			for _, dir := range []string{"Gospel", "Rock", "Pop", "Folk", "Old Songs"} {
				err := user.WSConn.WriteJSON(map[string]any{
					"action": "mkdir",
					"data": map[string]any{
						"parentDirectoryId": nativeRootDirs["Music"],
						"directoryName":     dir,
					},
				})

				require.NoError(t, err)

				serverWSReply := <-user.ServerWSMsg

				td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
					"event":    "server reply",
					"toAction": "mkdir",
					"data": td.SuperMapOf(map[string]any{
						"id":       td.Ignore(),
						"obj_type": "directory",
						"name":     dir,
					}, nil),
				}, nil))
			}
		}

		musicDirs := make(map[string]string, 5)

		{
			t.Log("list the dirs in native dir: 'Music'")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "ls",
				"data": map[string]any{
					"directoryId": nativeRootDirs["Music"],
				},
			})

			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "ls",
				"data":     containsDirs("Gospel", "Rock", "Pop", "Folk", "Old Songs"),
			}, nil))

			for _, dm := range serverWSReply["data"].([]any) {
				m := dm.(map[string]any)
				musicDirs[m["name"].(string)] = m["id"].(string)
			}
		}

		{
			t.Log("trash 'Folk' and 'Old Songs' dirs in native dir: 'Music'")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "trash",
				"data": map[string]any{
					"parentDirectoryId": nativeRootDirs["Music"],
					"objectIds":         []string{musicDirs["Folk"], musicDirs["Old Songs"]},
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "trash",
				"data":     true,
			}, nil))
		}

		{
			t.Log("list the dirs now in native dir: 'Music' | confirm trashing")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "ls",
				"data": map[string]any{
					"directoryId": nativeRootDirs["Music"],
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "ls",
				"data":     td.All(containsDirs("Gospel", "Rock", "Pop"), notContainsDirs("Folk", "Old Songs")),
			}, nil))

			clear(musicDirs)

			for _, dm := range serverWSReply["data"].([]any) {
				m := dm.(map[string]any)
				musicDirs[m["name"].(string)] = m["id"].(string)
			}
		}

		trashDirs := make(map[string]string)

		{
			t.Log("view dirs in Trash")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "view trash",
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "view trash",
				"data":     containsDirs("Folk", "Old Songs"),
			}, nil))

			for _, dm := range serverWSReply["data"].([]any) {
				m := dm.(map[string]any)
				trashDirs[m["name"].(string)] = m["id"].(string)
			}
		}

		{
			t.Log("restore 'Folk' dir from Trash")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "restore",
				"data": map[string]any{
					"objectIds": []string{trashDirs["Folk"]},
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "restore",
				"data":     true,
			}, nil))
		}

		{
			t.Log("view dirs now in Trash | confirm 'Folk' dir restored")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "view trash",
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "view trash",
				"data":     td.All(notContainsDirs("Folk"), containsDirs("Old Songs")),
			}, nil))

			for _, dm := range serverWSReply["data"].([]any) {
				m := dm.(map[string]any)
				trashDirs[m["name"].(string)] = m["id"].(string)
			}
		}

		{
			t.Log("list the dirs now in native dir: 'Music' | confirm 'Folk' dir restored")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "ls",
				"data": map[string]any{
					"directoryId": nativeRootDirs["Music"],
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "ls",
				"data":     td.All(containsDirs("Gospel", "Rock", "Pop", "Folk"), notContainsDirs("Old Songs")),
			}, nil))

			clear(musicDirs)

			for _, dm := range serverWSReply["data"].([]any) {
				m := dm.(map[string]any)
				musicDirs[m["name"].(string)] = m["id"].(string)
			}
		}

		{
			t.Log("attempt to trash a native directory fails")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "trash",
				"data": map[string]any{
					"parentDirectoryId": "/",
					"objectIds":         []string{nativeRootDirs["Documents"]},
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "trash",
				"data":     false,
			}, nil))
		}

		{
			t.Log("rename 'Gospel' dir in native root dir: 'Music' to 'Christian Music'")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "rename",
				"data": map[string]any{
					"parentDirectoryId": nativeRootDirs["Music"],
					"objectId":          musicDirs["Gospel"],
					"newName":           "Christian Music",
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "rename",
				"data":     true,
			}, nil))
		}

		{
			t.Log("list the dirs now in native dir: 'Music' | confirm renaming")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "ls",
				"data": map[string]any{
					"directoryId": nativeRootDirs["Music"],
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "ls",
				"data":     td.All(containsDirs("Christian Music", "Rock", "Pop", "Folk"), notContainsDirs("Gospel")),
			}, nil))

			clear(musicDirs)

			for _, dm := range serverWSReply["data"].([]any) {
				m := dm.(map[string]any)
				musicDirs[m["name"].(string)] = m["id"].(string)
			}
		}

		{
			t.Log("attempt to rename a native directory fails")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "rename",
				"data": map[string]any{
					"parentDirectoryId": "/",
					"objectId":          nativeRootDirs["Pictures"],
					"newName":           "Images",
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "rename",
				"data":     false,
			}, nil))
		}
	}
}
