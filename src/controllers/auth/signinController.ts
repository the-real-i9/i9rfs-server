import type { Request, Response } from "express"
import * as signinService from "../../services/appServices/signinService.ts"
import { StatusCodes } from "http-status-codes"

export async function Signin(req: Request, res: Response) {
  try {
    const body: { emailOrUsername: string; password: string } = req.body

    const { respData, authJwt } = await signinService.Signin(
      body.emailOrUsername,
      body.password
    )

    req.sessionOptions.expires = new Date(Date.now() + 10 * 24 * 60 * 60 * 1000) // 10 days

    req.session = {
      user: { authJwt },
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
