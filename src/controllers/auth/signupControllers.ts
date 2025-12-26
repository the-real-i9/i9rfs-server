import type { Request, Response } from "express"
import * as signupService from "../../services/appServices/signupService.ts"
import { StatusCodes } from "http-status-codes"

export async function RequestNewAccount(req: Request, res: Response) {
  try {
    const body: { email: string } = req.body

    const { respData, sessionData } = await signupService.RequestNewAccount(
      body.email
    )

    req.sessionOptions.expires = new Date(Date.now() + 60 * 60 * 1000)

    req.session = {
      signup: JSON.stringify(sessionData),
    }

    return res.json(respData)
  } catch (error: any) {
    if (error.name === "AppError") {
      return res.status(error.code).json(error.message)
    }

    console.error(error)
    return res.sendStatus(StatusCodes.INTERNAL_SERVER_ERROR)
  }
}

export async function VerifyEmail(req: Request, res: Response) {
  try {
    const sessionData = res.locals.signup_sess_data

    const body: { code: string } = req.body

    const { respData, newSessionData } = await signupService.VerifyEmail(
      sessionData,
      body.code
    )

    req.sessionOptions.expires = new Date(Date.now() + 60 * 60 * 1000)

    req.session = {
      signup: JSON.stringify(newSessionData),
    }

    return res.json(respData)
  } catch (error: any) {
    if (error.name === "AppError") {
      return res.status(error.code).json(error.message)
    }

    console.error(error)
    return res.sendStatus(StatusCodes.INTERNAL_SERVER_ERROR)
  }
}

export async function RegisterUser(req: Request, res: Response) {
  try {
    const sessionData = res.locals.signup_sess_data

    const body: { username: string; password: string } = req.body

    const { respData, authJwt } = await signupService.RegisterUser(
      sessionData,
      body.username,
      body.password
    )

    req.sessionOptions.expires = new Date(Date.now() + 10 * 24 * 60 * 60 * 1000) // 10 days

    req.session = {
      user: JSON.stringify({ authJwt }),
    }

    return res.status(StatusCodes.CREATED).json(respData)
  } catch (error: any) {
    if (error.name === "AppError") {
      return res.status(error.code).json(error.message)
    }

    console.error(error)
    return res.sendStatus(StatusCodes.INTERNAL_SERVER_ERROR)
  }
}
