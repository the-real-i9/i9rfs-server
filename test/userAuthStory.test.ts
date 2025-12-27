import assert from "node:assert"
import { beforeEach, afterEach, test } from "node:test"
import request from "superwstest"
import { StatusCodes } from "http-status-codes"

import server from "../src/index.ts"

const signupPath = "/api/auth/signup"
const signinPath = "/api/auth/signin"

const signoutPath = "/api/app/signout"

const rfsPath = "/rfs"

beforeEach((_, done) => {
  server.listen(0, "localhost", done)
})

afterEach((_, done) => {
  server.close(done)
})

test("TestUserAuthStory", async () => {
  const user = {
    email: "suberu@gmail.com",
    username: "suberu",
    password: "sketeppy",
    sessionCookie: "",
  }

  {
    console.log("Action: user requests new account")

    const res = await request(server)
      .post(signupPath + "/request_new_account")
      .send({ email: user.email })
      .set("Accept", "application/json")
      .expect("Content-Type", /json/)

    if (res.statusCode !== StatusCodes.OK) {
      console.error("unexpected error:", res.body)
    }

    assert.equal(res.statusCode, StatusCodes.OK)
    assert.partialDeepStrictEqual(res.body, {
      msg: `Enter the 6-digit code sent to ${user.email} to verify your email`,
    })

    user.sessionCookie = res.header["set-cookie"]
  }

  {
    console.log("Action: user sends an incorrect email verf code")

    const res = await request(server)
      .post(signupPath + "/verify_email")
      .send({ code: "011111" })
      .set("Cookie", user.sessionCookie)
      .set("Accept", "application/json")
      .expect("Content-Type", /json/)

    if (res.statusCode !== StatusCodes.BAD_REQUEST) {
      console.error("unexpected error:", res.body)
    }

    assert.equal(res.statusCode, StatusCodes.BAD_REQUEST)
    assert.equal(
      res.body,
      "Incorrect verification code! Check or Re-submit your email."
    )
  }

  {
    console.log("Action: user sends the correct email verification code")

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

    assert.equal(res.statusCode, StatusCodes.OK)
    assert.partialDeepStrictEqual(res.body, {
      msg: `Your email, ${user.email}, has been verified!`,
    })

    user.sessionCookie = res.header["set-cookie"]
  }

  {
    console.log("Action: user submits her information")

    const res = await request(server)
      .post(signupPath + "/register_user")
      .send({ username: user.username, password: user.password })
      .set("Cookie", user.sessionCookie)
      .set("Accept", "application/json")
      .expect("Content-Type", /json/)

    if (res.statusCode !== StatusCodes.CREATED) {
      console.error("unexpected error:", res.body)
    }

    assert.equal(res.statusCode, StatusCodes.CREATED)
    assert.ok(
      res.body?.user?.username,
      "user.username doesn't exist on res.body"
    )
    assert.equal(res.body?.msg, "Signup success!")

    user.sessionCookie = res.header["set-cookie"]
  }

  {
    console.log("Action: user signs out")

    const res = await request(server)
      .get(signoutPath)
      .set("Cookie", user.sessionCookie)

    if (res.statusCode !== StatusCodes.OK) {
      console.error("unexpected error:", res.body)
    }

    assert.equal(res.statusCode, StatusCodes.OK)
  }

  {
    console.log("Action: user signs in with incorrect credentials")

    const res = await request(server)
      .post(signinPath)
      .send({
        emailOrUsername: user.email,
        password: "millinix",
      })
      .set("Accept", "application/json")
      .expect("Content-Type", /json/)

    if (res.statusCode !== StatusCodes.NOT_FOUND) {
      console.error("unexpected error:", res.body)
    }

    assert.equal(res.statusCode, StatusCodes.NOT_FOUND)
    assert.equal(res.body, "Incorrect email or password")
  }

  {
    console.log("Action: user signs in with correct credentials")

    const res = await request(server)
      .post(signinPath)
      .send({
        emailOrUsername: user.email,
        password: user.password,
      })
      .set("Accept", "application/json")
      .expect("Content-Type", /json/)

    if (res.statusCode !== StatusCodes.OK) {
      console.error("unexpected error:", res.body)
    }

    assert.equal(res.statusCode, StatusCodes.OK)
    assert.equal(res.body.msg, "Signin success!")

    user.sessionCookie = res.header["set-cookie"]
  }
})
