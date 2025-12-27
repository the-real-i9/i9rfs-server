import type { Response, Request, NextFunction } from "express"
import { StatusCodes } from "http-status-codes"
import * as securityServices from "../services/appServices/securityServices.ts"

export function SignupSession(req: Request, res: Response, next: NextFunction) {
  try {
    const ssStr: string = req.session?.signup || ""

    if (!ssStr) {
      return res
        .status(StatusCodes.UNAUTHORIZED)
        .json("no ongoing signup session")
    }

    res.locals.signup_sess_data = JSON.parse(ssStr)

    return next()
  } catch (error) {
    return next(error)
  }
}

export function UserAuth(req: Request, res: Response, next: NextFunction) {
  try {
    const usStr: string = req.session?.user || ""

    if (!usStr) {
      return res
        .status(StatusCodes.UNAUTHORIZED)
        .json("authentication required")
    }

    const sessionToken = JSON.parse(usStr).authJwt

    const user = securityServices.JwtVerify(
      sessionToken,
      process.env.AUTH_JWT_SECRET || ""
    )

    res.locals.user = user

    return next()
  } catch (error: any) {
    if (error.name === "TokenExpiredError") {
      return res.status(StatusCodes.UNAUTHORIZED).json("session expired")
    }

    return next(error)
  }
}
