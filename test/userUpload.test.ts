import fs from "node:fs/promises"
import path from "node:path"
import { beforeEach, afterEach, test, type TestContext } from "node:test"
import request from "superwstest"
import { StatusCodes } from "http-status-codes"

import {
  logProgress,
  startResumableUpload,
  uploadFileInChunks,
} from "./testHelpers.ts"
import server from "../src/index.ts"
import type { FileT } from "../src/appTypes.ts"

const signupPath = "/api/auth/signup"
const uploadPath = "/api/app/uploads"

beforeEach((_, done) => {
  server.listen(0, "localhost", done)
})

afterEach((_, done) => {
  server.close(done)
})

test("TestUserFileUpload", async (t: TestContext) => {
  const user = {
    email: "louislitt@gmail.com",
    username: "louislitt",
    password: "pearsonsuckman",
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

  /* ---------------- */
  {
    console.log("Action: upload file: authorize upload")

    const filePath = path.resolve("./test/test_files/Aye Ole - Infinity.mp3")
    const contentType = "audio/mp3"

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

    user.sessionCookie = res.header["set-cookie"]

    const { uploadUrl, objectId, cloudObjectName } = res.body

    const sessionUrl = await startResumableUpload(uploadUrl, contentType)

    console.log("Resumable session started:")

    await uploadFileInChunks(sessionUrl, filePath, contentType, logProgress)

    console.log("Upload complete")

    {
      console.log("Action: upload file: cloud upload complete")

      const res = await request(server)
        .post(uploadPath + "/cloud_upload_complete")
        .send({ cloudObjectName })
        .set("Cookie", user.sessionCookie)
        .set("Accept", "application/json")
        .expect("Content-Type", /json/)

      if (res.statusCode !== StatusCodes.OK) {
        console.error("unexpected error:", res.body)
      }

      t.assert.strictEqual(res.statusCode, StatusCodes.OK)
      t.assert.strictEqual(res.body, true)

      user.sessionCookie = res.header["set-cookie"]
    }

    {
      console.log("Action: upload file: create file object")

      const res = await request(server)
        .post(uploadPath + "/create_file_object")
        .send({
          parentDirectoryId: "/",
          objectId,
          cloudObjectName,
          displayName: "Aye-Ole.mp3",
        })
        .set("Cookie", user.sessionCookie)
        .set("Accept", "application/json")
        .expect("Content-Type", /json/)

      if (res.statusCode !== StatusCodes.OK) {
        console.error("unexpected error:", res.body)
      }

      t.assert.strictEqual(res.statusCode, StatusCodes.OK)
      t.assert.ok(res.body)

      const data: FileT = res.body

      t.assert.ok(data.id)
      t.assert.partialDeepStrictEqual(res.body, {
        obj_type: "file",
        name: "Aye-Ole.mp3",
        cloud_object_name: cloudObjectName,
        mime_type: contentType,
      })

      user.sessionCookie = res.header["set-cookie"]
    }
  }
})
