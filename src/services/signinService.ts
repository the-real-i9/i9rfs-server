import { StatusCodes } from "http-status-codes"
import * as user from "../models/userModel.ts"
import * as securityServices from "./util/securityServices.ts"
import type { ClientUserT } from "../appTypes.ts"

interface SigninResp {
  msg: string
  user: ClientUserT
}

export async function Signin(emailOrUsername: string, inputPassword: string) {
  const theUser = await user.AuthFind(emailOrUsername)

  if (!theUser) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_FOUND,
      message: "Incorrect email or password",
    }
  }

  const yes = await securityServices.PasswordMatchesHash(
    theUser.password,
    inputPassword
  )

  if (!yes) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_FOUND,
      message: "Incorrect email or password",
    }
  }

  const authJwt = securityServices.JwtSign(
    {
      username: theUser.username,
    },
    process.env.AUTH_JWT_SECRET || "",
    10 * 24 * 60 * 60 * 1000
  ) // 10 days

  const respData: SigninResp = {
    msg: "Signin success!",
    user: { username: theUser.username },
  }

  return { respData, authJwt }
}
