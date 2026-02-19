import path from "node:path"
import fs from "node:fs/promises"
import request from "superwstest"
import { StatusCodes } from "http-status-codes"
import { beforeEach, afterEach, test, type TestContext } from "node:test"

import server from "../src/index.ts"
import {
  containsDirs,
  containsFiles,
  logProgress,
  notContainsDirs,
  startResumableUpload,
  uploadFileInChunks,
} from "./testHelpers.ts"
import type { FileT, DirT } from "../src/appTypes.ts"
import appGlobals from "../src/appGlobals.ts"

const signupPath = "/api/auth/signup"
const uploadPath = "/api/app/uploads"
const wsPath = "/ws"

beforeEach((_, done) => {
  server.listen(0, "localhost", done)
})

afterEach((_, done) => {
  server.close(done)
})

test("TestUserRFSActionStory", async (t: TestContext) => {
  t.skip()
  if (1 === 1) {
    return
  }

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
      .ws(wsPath)
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
      .ws(wsPath)
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
      .ws(wsPath)
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
      .ws(wsPath)
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
      .ws(wsPath)
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
      .ws(wsPath)
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
      .ws(wsPath)
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
      .ws(wsPath)
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
      .ws(wsPath)
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
      .ws(wsPath)
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

  {
    console.log("trash 'Folk' and 'Old Songs' dirs in native dir: 'Music'")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "trash",
        data: {
          parentDirectoryId: nativeRootDirs.Music,
          objectIds: [musicDirs.Folk, musicDirs["Old Songs"]],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "trash",
          data: true,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: list the dirs now in native dir: 'Music' | confirm trashing"
    )

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "ls",
        data: {
          directoryId: nativeRootDirs.Music,
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "ls",
        })
        containsDirs(msg.data as DirT[], ["Gospel", "Rock", "Pop"], t)
        notContainsDirs(msg.data as DirT[], ["Folk", "Old Songs"], t)

        musicDirs = {}

        for (const dir of msg.data as DirT[]) {
          musicDirs[dir.name] = dir.id
        }

        return true
      })
      .close()
      .expectClosed()
  }

  let trashDirs: { [x: string]: string } = {}

  {
    console.log("view dirs in Trash")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "viewtrash",
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "viewtrash",
        })
        containsDirs(msg.data as DirT[], ["Folk", "Old Songs"], t)

        for (const dir of msg.data as DirT[]) {
          trashDirs[dir.name] = dir.id
        }

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log("restore 'Folk' dir from Trash")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "restore",
        data: {
          objectIds: [trashDirs.Folk],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "restore",
          data: true,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log("view dirs now in Trash | confirm 'Folk' dir restored")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "viewtrash",
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "viewtrash",
        })
        notContainsDirs(msg.data as DirT[], ["Folk"], t)
        containsDirs(msg.data as DirT[], ["Old Songs"], t)

        trashDirs = {}

        for (const dir of msg.data as DirT[]) {
          trashDirs[dir.name] = dir.id
        }

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: list the dirs now in native dir: 'Music' | confirm 'Folk' dir restored"
    )

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "ls",
        data: {
          directoryId: nativeRootDirs.Music,
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "ls",
        })
        containsDirs(msg.data as DirT[], ["Gospel", "Rock", "Pop", "Folk"], t)
        notContainsDirs(msg.data as DirT[], ["Old Songs"], t)

        musicDirs = {}

        for (const dir of msg.data as DirT[]) {
          musicDirs[dir.name] = dir.id
        }

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log("attempt to trash a native directory fails")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "trash",
        data: {
          parentDirectoryId: "/",
          objectIds: [nativeRootDirs.Documents],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "trash",
          data: false,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "rename 'Gospel' dir in native root dir: 'Music' to 'Christian Music'"
    )

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "rename",
        data: {
          parentDirectoryId: nativeRootDirs.Music,
          objectId: musicDirs.Gospel,
          newName: "Christian Music",
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "rename",
          data: true,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: list the dirs now in native dir: 'Music' | confirm renaming"
    )

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "ls",
        data: {
          directoryId: nativeRootDirs.Music,
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "ls",
        })
        containsDirs(
          msg.data as DirT[],
          ["Christian Music", "Rock", "Pop", "Folk"],
          t
        )
        notContainsDirs(msg.data as DirT[], ["Gospel"], t)

        musicDirs = {}

        for (const dir of msg.data as DirT[]) {
          musicDirs[dir.name] = dir.id
        }

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log("attempt to rename a native directory fails")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "rename",
        data: {
          parentDirectoryId: "/",
          objectId: nativeRootDirs.Pictures,
          newName: "Images",
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "rename",
          data: false,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: create nested sub-directories in 'Rock' dir | to confirm whole branch move"
    )

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "mkdir",
        data: {
          parentDirectoryId: musicDirs.Rock,
          directoryNames: ["Pop Rock/Afro Pop Rock"],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "mkdir",
        })
        containsDirs(msg.data as DirT[], ["Pop Rock"], t)

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: move 'Rock' and 'Pop' dirs from/to native root dir 'Music'/'Downloads'"
    )

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "move",
        data: {
          fromParentDirectoryId: nativeRootDirs.Music,
          toParentDirectoryId: nativeRootDirs.Downloads,
          objectIds: [musicDirs.Rock, musicDirs.Pop],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "move",
          data: true,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log(
      "Action: list the dirs in native dir: 'Downloads' and 'Music' | confirm move"
    )

    await request(server)
      .ws(wsPath)
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
        containsDirs(msg.data as DirT[], ["Pop", "Rock"], t)

        return true
      })
      .sendJson({
        command: "ls",
        data: {
          directoryId: nativeRootDirs.Music,
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "ls",
        })
        notContainsDirs(msg.data as DirT[], ["Pop", "Rock"], t)

        return true
      })
      .close()
      .expectClosed()
  }

  /* ==== FILE TESTING ==== */

  let uploadUrl: string, fileObjectId: string, cloudObjectName: string
  const filePath = path.resolve("./test/test_files/Aye Ole - Infinity.mp3")
  const contentType = "audio/mp3"

  {
    console.log("Action: upload file: authorize upload")

    const res = await request(server)
      .post(uploadPath + "/authorize")
      .send({ mimeType: contentType, size: (await fs.stat(filePath)).size })
      .set("Cookie", user.sessionCookie)
      .set("Accept", "application/json")
      .expect("Content-Type", /json/)

    if (res.statusCode !== StatusCodes.OK) {
      console.error("unexpected error:", res.body)
    }

    t.assert.strictEqual(res.statusCode, StatusCodes.OK)
    t.assert.ok(res.body.uploadUrl)
    t.assert.ok(res.body.objectId)
    t.assert.ok(res.body.cloudObjectName)

    uploadUrl = res.body.uploadUrl
    fileObjectId = res.body.objectId
    cloudObjectName = res.body.cloudObjectName
  }

  {
    console.log("Upload session started:")

    const sessionUrl = await startResumableUpload(uploadUrl, contentType)

    await uploadFileInChunks(sessionUrl, filePath, contentType, logProgress)

    console.log("Upload complete")
  }

  {
    console.log("Action: create file in root dir")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "mkfil",
        data: {
          parentDirectoryId: "/",
          objectId: fileObjectId,
          cloudObjectName,
          filename: "Aye-Ole.mp3",
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "mkfil",
        })

        const data: FileT = msg.data
        t.assert.ok(data.id)
        t.assert.partialDeepStrictEqual(data, {
          obj_type: "file",
          name: "Aye-Ole.mp3",
          cloud_object_name: cloudObjectName,
          mime_type: contentType,
        })

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log("Action: list files in root")

    await request(server)
      .ws(wsPath)
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
        containsFiles(msg.data as FileT[], ["Aye-Ole.mp3"], t)

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log("Action: download file created in root")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "download",
        data: {
          fileObjectId,
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "download",
        })
        t.assert.ok(typeof msg.data === "string")

        const msgData: string = msg.data
        t.assert.ok(msgData.startsWith("https://"))

        return true
      })
      .close()
      .expectClosed()
  }

  let oldMusicDirId: string = ""

  {
    console.log("Action: create 'Old Music' dir in 'Music' dir")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "mkdir",
        data: {
          parentDirectoryId: nativeRootDirs.Music,
          directoryNames: ["Old Music"],
        },
      })
      .expectJson((msg) => {
        t.assert.partialDeepStrictEqual(msg, {
          event: "server reply",
          toCommand: "mkdir",
        })
        const msgData: DirT[] = msg.data
        containsDirs(msgData, ["Old Music"], t)

        oldMusicDirId = msgData[0].id

        return true
      })
      .close()
      .expectClosed()
  }

  {
    console.log("Action: copy file in root to 'Old Music' dir in 'Music'")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "copy",
        data: {
          fromParentDirectoryId: "/",
          toParentDirectoryId: oldMusicDirId,
          objectIds: [fileObjectId],
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
    console.log("Action: delete file in root and 'Old Music' dir")

    await request(server)
      .ws(wsPath)
      .set("Cookie", user.sessionCookie)
      .sendJson({
        command: "del",
        data: {
          parentDirectoryId: "/",
          objectIds: [fileObjectId],
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
      .sendJson({
        command: "del",
        data: {
          parentDirectoryId: nativeRootDirs.Music,
          objectIds: [oldMusicDirId],
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
    console.log("Action: cleanup bucket")
    await appGlobals.AppGCSBucket.deleteFiles()
  }
})
