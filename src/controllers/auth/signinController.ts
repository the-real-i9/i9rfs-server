import type { NextFunction, Request, Response } from "express"
import * as signinService from "../../services/signinService.ts"
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

export function SigninDoc(req: Request, res: Response, next: NextFunction) {
  /*
  #swagger.summary = 'Signin user'
  #swagger.description = 'Signin with email/username and password'
  #swagger.tags = ['auth']
  #swagger.requestBody = {
    required: true,
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            emailOrUsername: {
              oneOf: [
                { type: "string", format: "email" },
                { type: "string", minLength: 3 }
              ]
            },
            password: {
              type: "string",
              minLength: 8
            }
          },
          required: ["emailOrUsername", "password"]
        }
      }
    }
  } 

  #swagger.responses[200] = {
    description: 'Signin success',
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            msg: {
              type: "string"
            },
            user: {
              $ref: "#/components/schemas/ClientUserT"
            }
          }
        }
      }
    },
    headers: {
      "Set-Cookie": {
        description: 'Auth user session cookie',
        schema: {
          type: "string"
        }
      }
    }
  } 

  #swagger.responses[400] = {
    description: 'Incorrect email/username or password',
    content: {
      "application/json": {
        schema: {
          type: "string"
        }
      }
    }
  } 

  #swagger.responses[500] = {
    description: 'Internal Server Error',
    content: {
      "application/json": {
        schema: {
          type: "string"
        }
      }
    }
  } 
  */
  next()
}
