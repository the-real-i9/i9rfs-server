import type { Request, Response } from "express"
import * as signupService from "../../services/signupService.ts"
import { StatusCodes } from "http-status-codes"

export async function RequestNewAccount(req: Request, res: Response) {
  /*
  #swagger.summary = 'Signup user - Step 1'
  #swagger.description = 'Submit email to request a new account'
  #swagger.tags = ['auth']
  #swagger.requestBody = {
    required: true,
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            email: {
              type: "string",
              format: "email",
              example: "example@gmail.com"
            }
          },
          required: ["email"]
        }
      }
    }
  } 

  #swagger.responses[200] = {
    description: 'Step 1 success: Proceed to email verification',
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            msg: {
              type: "string",
            }
          }
        }
      }
    }
  } 

  #swagger.responses[400] = {
    description: 'Step 1 failed: A user with this email already exists',
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            msg: {
              type: "string",
            }
          }
        }
      }
    }
  } 

  #swagger.responses[500] = {
    description: 'Step 1 failed: Internal Server Error',
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            msg: {
              type: "string",
            }
          }
        }
      }
    }
  } 
  */

  try {
    const body: { email: string } = req.body

    const { respData, sessionData } = await signupService.RequestNewAccount(
      body.email
    )

    req.sessionOptions.expires = new Date(Date.now() + 60 * 60 * 1000)

    req.session = {
      signup: sessionData,
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
      signup: newSessionData,
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
      user: { authJwt },
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
