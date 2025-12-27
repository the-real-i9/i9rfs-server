import { beforeEach, afterEach, test, type TestContext } from "node:test"
import request from "superwstest"
import { StatusCodes } from "http-status-codes"

import server from "../src/index.ts"
import { containsDirs, notContainsDirs } from "./testHelpers.ts"
import { type DirT } from "../src/appTypes.ts"

const signupPath = "/api/auth/signup"

const rfsPath = "/rfs"

beforeEach((_, done) => {
  server.listen(0, "localhost", done)
})

afterEach((_, done) => {
  server.close(done)
})

test("TestUserRFSActionStory", async (t: TestContext) => {
  const user = {
    email: "mikeross@gmail.com",
    username: "mikeross",
    password: "paralegal_zane",
    sessionCookie: "",
  }

  console.log("Action: user creates new account")

  {
    const res = await request(server)
      .post(signupPath + "/request_new_account")
      .send({ email: user.email })
      .set("Accept", "application/json")
      .expect("Content-Type", /json/)

    if (res.statusCode !== StatusCodes.OK) {
      console.error("unexpected error:", res.body)
    }

    t.assert.strictEqual(res.statusCode, StatusCodes.OK)
    t.assert.partialDeepStrictEqual(res.body, {
      msg: `Enter the 6-digit code sent to ${user.email} to verify your email`,
    })

    user.sessionCookie = res.header["set-cookie"]
  }

  {
    const verfCode = process.env.DUMMY_TOKEN

    const res = await request(server)
      .post(signupPath + "/verify_email")
      .send({ code: verfCode })
      .set("Cookie", user.sessionCookie)
      .set("Accept", "application/json")
      .expect("Content-Type", /json/)

    if (res.statusCode !== StatusCodes.OK) {
      console.error("unexpected error:", res.body)
    }

    t.assert.strictEqual(res.statusCode, StatusCodes.OK)
    t.assert.partialDeepStrictEqual(res.body, {
      msg: `Your email, ${user.email}, has been verified!`,
    })

    user.sessionCookie = res.header["set-cookie"]
  }

  {
    const res = await request(server)
      .post(signupPath + "/register_user")
      .send({ username: user.username, password: user.password })
      .set("Cookie", user.sessionCookie)
      .set("Accept", "application/json")
      .expect("Content-Type", /json/)

    if (res.statusCode !== StatusCodes.CREATED) {
      console.error("unexpected error:", res.body)
    }

    t.assert.strictEqual(res.statusCode, StatusCodes.CREATED)
    t.assert.ok(
      res.body?.user?.username,
      "user.username doesn't exist on res.body"
    )
    t.assert.strictEqual(res.body?.msg, "Signup success!")

    user.sessionCookie = res.header["set-cookie"]
  }

  const nativeRootDirs: { [x: string]: string } = {}

  {
    console.log("Action: list native directories in root")

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "ls",
        data: {
          directoryId: "/",
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "ls",
        })
        containsDirs(
          msg.data as DirT[],
          ["Documents", "Downloads", "Music", "Videos", "Pictures"],
          t
        )

        for (const dir of msg.data as DirT[]) {
          nativeRootDirs[dir.name] = dir.id
        }

        return true
      })
      .close()
      .expectClosed()
  }

  let videoDirs: { [x: string]: string } = {}

  {
    console.log("Action: bulk create dirs in native dir: 'Videos'")

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "mkdir",
        data: {
          parentDirectoryId: nativeRootDirs.Videos,
          directoryNames: [
            "Horror",
            "Comedy",
            "Legal",
            "Musical",
            "Action",
            "NotAVideo",
            "DeleteMe",
          ],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "mkdir",
        })
        containsDirs(
          msg.data as DirT[],
          [
            "Horror",
            "Comedy",
            "Legal",
            "Musical",
            "Action",
            "NotAVideo",
            "DeleteMe",
          ],
          t
        )

        for (const dir of msg.data as DirT[]) {
          videoDirs[dir.name] = dir.id
        }

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: put a sub-directory inside 'DeleteMe' dir | to test recursive deletion"
    )

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "mkdir",
        data: {
          parentDirectoryId: videoDirs.DeleteMe,
          directoryNames: ["DeleteMyChild"],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "mkdir",
        })
        containsDirs(msg.data as DirT[], ["DeleteMyChild"], t)

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: delete 'NotAVideo' and 'DeleteMe' dirs in native root dir: 'Videos'"
    )

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "del",
        data: {
          parentDirectoryId: nativeRootDirs.Videos,
          objectIds: [videoDirs.NotAVideo, videoDirs.DeleteMe],
        },
      })
      .expectJson((msg) => {
        t.assert.deepStrictEqual(msg, {
          event: "server reply",
          toCommand: "del",
          data: true,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: list the dirs now in native dir: 'Videos' | confirm deletion"
    )

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "ls",
        data: {
          directoryId: nativeRootDirs.Videos,
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "ls",
        })
        containsDirs(
          msg.data as DirT[],
          ["Horror", "Comedy", "Legal", "Musical", "Action"],
          t
        )
        notContainsDirs(msg.data as DirT[], ["NotAVideo", "DeleteMe"], t)

        videoDirs = {}

        for (const dir of msg.data as DirT[]) {
          videoDirs[dir.name] = dir.id
        }

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log("attempt to delete a native directory fails")

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "del",
        data: {
          parentDirectoryId: "/",
          objectIds: [nativeRootDirs.Downloads],
        },
      })
      .expectJson((msg) => {
        t.assert.deepStrictEqual(msg, {
          event: "server reply",
          toCommand: "del",
          data: false,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: put sub-directories inside 'Horror' dir | to test recursive copy"
    )

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "mkdir",
        data: {
          parentDirectoryId: videoDirs.Horror,
          directoryNames: ["The Conjuring/Season 1/Episodes"],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "mkdir",
        })
        containsDirs(msg.data as DirT[], ["The Conjuring"], t)

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: copy 'Horror' and 'Comedy' dirs from/to native root dir 'Videos'/'Downloads'"
    )

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "copy",
        data: {
          fromParentDirectoryId: nativeRootDirs.Videos,
          toParentDirectoryId: nativeRootDirs.Downloads,
          objectIds: [videoDirs.Horror, videoDirs.Comedy],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "copy",
          data: true,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: list the dirs in native dir: 'Downloads' | confirm copy"
    )

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "ls",
        data: {
          directoryId: nativeRootDirs.Downloads,
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "ls",
        })
        containsDirs(msg.data as DirT[], ["Horror", "Comedy"], t)

        return true
      })
      .close()
      .expectClosed()
  }

  let musicDirs: { [x: string]: string } = {}

  {
    console.log("Action: bulk create dirs in native dir: 'Music'")

    await request(server)
      .ws(rfsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "mkdir",
        data: {
          parentDirectoryId: nativeRootDirs.Music,
          directoryNames: ["Gospel", "Rock", "Pop", "Folk", "Old Songs"],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "mkdir",
        })
        containsDirs(
          msg.data as DirT[],
          ["Gospel", "Rock", "Pop", "Folk", "Old Songs"],
          t
        )

        for (const dir of msg.data as DirT[]) {
          musicDirs[dir.name] = dir.id
        }

        return true
      })
      .close()
      .expectClosed()
  }
})

/* func TestUserRFSActionStory(t *testing.T) {

	t.Log("----------")

	nativeRootDirs := make(map[string]string, 5)

	{

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
			t.Log("Action: list the dirs now in native dir: 'Music' | confirm trashing")

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
			t.Log("Action: list the dirs now in native dir: 'Music' | confirm 'Folk' dir restored")

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
			t.Log("Action: list the dirs now in native dir: 'Music' | confirm renaming")

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

		{
			t.Log("Action: create nested sub-directories in 'Rock' dir | to confirm whole branch move")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "mkdir",
				"data": map[string]any{
					"parentDirectoryId": musicDirs["Rock"],
					"directoryName":     "Pop Rock",
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
					"name":     "Pop Rock",
				}, nil),
			}, nil))

			err = user.WSConn.WriteJSON(map[string]any{
				"action": "mkdir",
				"data": map[string]any{
					"parentDirectoryId": serverWSReply["data"].(map[string]any)["id"].(string),
					"directoryName":     "Afro Pop Rock",
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
					"name":     "Afro Pop Rock",
				}, nil),
			}, nil))
		}

		{
			t.Log("Action: move 'Rock' and 'Pop' dirs from/to native root dir 'Music'/'Downloads'")

			err := user.WSConn.WriteJSON(map[string]any{
				"action": "move",
				"data": map[string]any{
					"fromParentDirectoryId": nativeRootDirs["Music"],
					"toParentDirectoryId":   nativeRootDirs["Downloads"],
					"objectIds":             []string{musicDirs["Rock"], musicDirs["Pop"]},
				},
			})
			require.NoError(t, err)

			serverWSReply := <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "move",
				"data":     true,
			}, nil))
		}

		{
			t.Log("Action: list the dirs in native dir: 'Downloads' and 'Music' | confirm move")

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
				"data":     containsDirs("Pop", "Rock"),
			}, nil))

			err = user.WSConn.WriteJSON(map[string]any{
				"action": "ls",
				"data": map[string]any{
					"directoryId": nativeRootDirs["Music"],
				},
			})
			require.NoError(t, err)

			serverWSReply = <-user.ServerWSMsg

			td.Cmp(td.Require(t), serverWSReply, td.Map(map[string]any{
				"event":    "server reply",
				"toAction": "ls",
				"data":     notContainsDirs("Pop", "Rock"),
			}, nil))
		}
	}
}
 */
