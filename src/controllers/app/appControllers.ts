import type { NextFunction, Request, Response } from "express"

import type { ClientUserT } from "../../appTypes.ts"
import { StatusCodes } from "http-status-codes"
import * as uploadService from "../../services/uploadService.ts"

export async function Signout(req: Request, res: Response) {
  const clientUser: ClientUserT = res.locals.user

  req.session = null

  return res.json(`Bye, ${clientUser.username}! See you again!`)
}

export async function AuthorizeUpload(req: Request, res: Response) {
  try {
    const clientUser: ClientUserT = res.locals.user

    const { mimeType, size }: { mimeType: string; size: number } = req.body

    const respData = await uploadService.AuthorizeUpload(
      clientUser.username,
      mimeType,
      size
    )

    return res.json(respData)
  } catch (error: any) {
    if (error.name === "AppError") {
      return res.status(error.code).json(error.message)
    }

    console.error(error)
    return res.sendStatus(StatusCodes.INTERNAL_SERVER_ERROR)
  }
}

/* === DOCS === */

export function SignoutDoc(req: Request, res: Response, next: NextFunction) {
  /*
  #swagger.summary = 'Signout user'
  #swagger.description = 'Signout user'
  #swagger.tags = ['app']

  #swagger.parameters['Cookie'] = {
    in: 'header',
    description: 'Auth user session cookie',
    required: true
  }

  #swagger.responses[200] = {
    description: 'Signout success',
    content: {
      "application/json": {
        schema: {
          type: "string"
        }
      }
    },
    headers: {
      "Set-Cookie": {
        description: 'Cookie invalidation',
        schema: {
          type: "string"
        }
      }
    }
  } 
  */
  next()
}

export function AuthorizeUploadDoc(
  req: Request,
  res: Response,
  next: NextFunction
) {
  /*
  #swagger.summary = 'Authorize file upload'
  #swagger.description = 'Authorize file upload'
  #swagger.tags = ['app']

  #swagger.parameters['Cookie'] = {
    in: 'header',
    description: 'Auth user session cookie',
    required: true
  }

  #swagger.requestBody = {
    required: true,
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            mimeType: {
              type: "string",
              format: "mime"
            },
            size: {
              type: "number",
              minimum: 1
            }
          },
          required: ["mimeType", "size"]
        }
      }
    }
  } 

  #swagger.responses[200] = {
    description: 'Upload authorization success',
    content: {
      "application/json": {
        schema: {
          type: "object",
          properties: {
            uploadUrl: {
              type: "string",
              format: "url"
            },
            objectId: {
              type: "string",
              format: "uuid"
            },
            cloudObjectName: {
              type: "string"
            }
          }
        }
      }
    }
  } 

  #swagger.responses[406] = {
    description: 'User does not have enough space',
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
