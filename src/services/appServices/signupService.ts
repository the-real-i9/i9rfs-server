import { StatusCodes } from "http-status-codes"
import * as user from "../../models/userModel.ts"
import * as securityServices from "./securityServices.ts"
import * as mailService from "../otherServices/mailService.ts"
import type { ClientUserT } from "../../appTypes.ts"

interface SignupRNASessionData {
  email: string
  vCode: string
  vCodeExpires: number
}

interface SignupRNAResp {
  msg: string
}

export async function RequestNewAccount(email: string) {
  if (await user.Exists(email)) {
    throw {
      name: "AppError",
      code: StatusCodes.BAD_REQUEST,
      message: "A user with this email already exists.",
    }
  }

  const { token: verfCode, expires } = securityServices.GetTokenAndExpiration()

  mailService.SendMail(
    email,
    "Verify your email",
    `<p>Your email verification code is <strong>${verfCode}</strong></p>`
  )

  const sessionData: SignupRNASessionData = {
    email: email,
    vCode: verfCode,
    vCodeExpires: expires,
  }

  const respData: SignupRNAResp = {
    msg: `Enter the 6-digit code sent to ${email} to verify your email`,
  }

  return { respData, sessionData }
}

interface SignupVESessionData {
  email: string
}

interface SignupVEResp {
  msg: string
}

export async function VerifyEmail(
  sd: SignupRNASessionData,
  inputVerfCode: string
) {
  if (sd.vCode != inputVerfCode) {
    throw {
      name: "AppError",
      code: StatusCodes.BAD_REQUEST,
      message: "Incorrect verification code! Check or Re-submit your email.",
    }
  }

  if (sd.vCodeExpires < Date.now()) {
    throw {
      name: "AppError",
      code: StatusCodes.BAD_REQUEST,
      message: "Verification code expired! Re-submit your email.",
    }
  }

  mailService.SendMail(
    sd.email,
    "Email Verification Success",
    `Your email <strong>${sd.email}</strong> has been verified!`
  )

  const newSessionData: SignupVESessionData = { email: sd.email }

  const respData: SignupVEResp = {
    msg: `Your email, ${sd.email}, has been verified!`,
  }

  return { respData, newSessionData }
}

interface SignupRUResp {
  msg: string
  user: ClientUserT
}

export async function RegisterUser(
  sessionData: SignupVESessionData,
  username: string,
  password: string
) {
  const { email } = sessionData

  if (await user.Exists(username)) {
    throw {
      name: "AppError",
      code: StatusCodes.BAD_REQUEST,
      message: "Username not available",
    }
  }

  const newUser = await user.New(
    email,
    username,
    await securityServices.HashPassword(password)
  )

  const authJwt = securityServices.JwtSign(
    newUser,
    process.env.AUTH_JWT_SECRET || "",
    10 * 24 * 60 * 60 * 1000
  ) // 10 days

  const respData: SignupRUResp = {
    msg: "Signup success!",
    user: newUser,
  }

  return { respData, authJwt }
}
