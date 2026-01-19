import type { NextFunction, Request, Response } from "express"
import * as signupService from "../../services/signupService.ts"
import { StatusCodes } from "http-status-codes"

export async function RequestNewAccount(req: Request, res: Response) {
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

/* ============ DOCS ============= */

export function RequestNewAccountDoc(
  req: Request,
  res: Response,
  next: NextFunction
) {
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
    description: 'Proceed to email verification',
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
    },
    headers: {
      "Set-Cookie": {
        description: 'Signup session cookie',
        schema: {
          type: "string"
        }
      }
    }
  } 

  #swagger.responses[400] = {
    description: 'A user with this email already exists',
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

export function VerifyEmailDoc(
  req: Request,
  res: Response,
  next: NextFunction
) {
  /*
  #swagger.summary = 'Signup user - Step 2'
  #swagger.description = 'Provide the 6-digit code sent to email'
  #swagger.tags = ['auth']
  
  #swagger.parameters['Cookie'] = {
    in: 'header',
    description: 'Signup session cookie',
    required: true
  }

  #swagger.requestBody = {
    required: true,
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            code: {
              type: "string",
              pattern: "^[0-9]{6}$",
              example: "123456"
            }
          },
          required: ["code"]
        }
      }
    }
  } 

  #swagger.responses[200] = {
    description: 'Email verified',
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
    },
    headers: {
      "Set-Cookie": {
        description: 'Signup session cookie',
        schema: {
          type: "string"
        }
      }
    }
  } 

  #swagger.responses[400] = {
    description: '(Incorrect verification code! | Verification code expired!) Check or Re-submit your email.',
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

export function RegisterUserDoc(
  req: Request,
  res: Response,
  next: NextFunction
) {
  /*
  #swagger.summary = 'Signup user - Step 3'
  #swagger.description = 'Provide remaining user credentials'
  #swagger.tags = ['auth']

  #swagger.parameters['Cookie'] = {
    in: 'header',
    description: 'Signup session cookie',
    required: true
  }

  #swagger.requestBody = {
    required: true,
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            username: {
              type: "string",
              minLength: 3
            },
            password: {
              type: "string",
              minLength: 8
            }
          },
          required: ["username", "password"]
        }
      }
    }
  } 

  #swagger.responses[200] = {
    description: 'Signup success',
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
    description: 'Username not available',
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
