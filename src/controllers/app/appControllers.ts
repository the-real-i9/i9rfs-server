import type { Request, Response } from "express"
import type { IncomingMessage } from "node:http"
import type WebSocket from "ws"

import type { ClientUserT } from "../../appTypes.ts"
import { GetSessionDataFromCookieHeader } from "../../helpers.ts"
import * as securityServices from "../../services/appServices/securityServices.ts"
import {
  copyCommandValid,
  delCommandValid,
  lsCommandValid,
  mkdirCommandValid,
  moveCommandValid,
  renameCommandValid,
  restoreCommandValid,
  rfsCommandBodyValid,
  trashCommandValid,
} from "./validation.ts"
import { StatusCodes } from "http-status-codes"
import * as helpers from "../../helpers.ts"
import * as rfsService from "../../services/appServices/rfsService.ts"
import * as uploadService from "../../services/appServices/uploadService.ts"

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

    if (req.session && typeof req.session === "object") {
      req.session.upload = {
        nextStep: "cloud-upload-complete",
        cloudObjectName: respData.cloudObjectName,
      }
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

export async function CloudUploadComplete(req: Request, res: Response) {
  try {
    const clientUser: ClientUserT = res.locals.user
    const { cloudObjectName }: { cloudObjectName: string } = req.body

    await uploadService.CloudUploadComplete(
      clientUser.username,
      cloudObjectName
    )

    if (req.session && typeof req.session === "object") {
      req.session.upload = {
        nextStep: "create-file-object",
        cloudObjectName,
      }
    }

    return res.json(true)
  } catch (error: any) {
    if (error.name === "AppError") {
      return res.status(error.code).json(error.message)
    }

    console.error(error)
    return res.sendStatus(StatusCodes.INTERNAL_SERVER_ERROR)
  }
}

export async function CreateFileObject(req: Request, res: Response) {
  try {
    const clientUser: ClientUserT = res.locals.user
    const data: {
      parentDirectoryId: string
      objectId: string
      cloudObjectName: string
      displayName: string
    } = req.body
    await rfsService.CreateFile(
      clientUser.username,
      data.parentDirectoryId,
      data.objectId,
      data.cloudObjectName,
      data.displayName
    )

    delete req.session?.upload

    return res.json(true)
  } catch (error: any) {
    if (error.name === "AppError") {
      return res.status(error.code).json(error.message)
    }

    console.error(error)
    return res.sendStatus(StatusCodes.INTERNAL_SERVER_ERROR)
  }
}

export function RFSController(ws: WebSocket, request: IncomingMessage) {
  try {
    const sessionData = GetSessionDataFromCookieHeader(
      request.headers.cookie || ""
    )
    if (!sessionData || !sessionData.user) {
      return ws.close(1001, "401: authentication required")
    }

    const {
      user: { authJwt },
    }: { user: { authJwt: string } } = sessionData

    const authPayload = securityServices.JwtVerify(
      authJwt,
      process.env.AUTH_JWT_SECRET || ""
    )

    if (typeof authPayload === "string") {
      throw new Error("invalid auth payload")
    }

    const { username: clientUser } = authPayload as ClientUserT

    ws.on("message", wsMessageHandler(ws, clientUser))
  } catch (error: any) {
    if (error.name === "TokenExpiredError") {
      return ws.close(1001, "401: session expired")
    }

    console.error(error)
    return ws.close(1001, "500: internal server error")
  }
}

function wsMessageHandler(ws: WebSocket, clientUsername: string) {
  return async function (data: WebSocket.RawData) {
    const rfsCommandBody = JSON.parse(data.toString())

    const { success, error } = rfsCommandBodyValid(rfsCommandBody)
    if (!success) {
      return ws.send(
        helpers.WSErrReply(
          StatusCodes.BAD_REQUEST,
          error,
          rfsCommandBody.command
        )
      )
    }

    const body = rfsCommandBody

    try {
      let resp

      switch (body.command) {
        case "ls": {
          const { success, error, data } = lsCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsService.Ls(clientUsername, data.directoryId)
          break
        }
        case "mkdir": {
          const { success, error, data } = mkdirCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsService.Mkdir(
            clientUsername,
            data.parentDirectoryId,
            data.directoryNames
          )
          break
        }
        case "del": {
          const { success, error, data } = delCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsService.Del(
            clientUsername,
            data.parentDirectoryId,
            data.objectIds
          )

          break
        }
        case "trash": {
          const { success, error, data } = trashCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsService.Trash(
            clientUsername,
            data.parentDirectoryId,
            data.objectIds
          )

          break
        }
        case "restore": {
          const { success, error, data } = restoreCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsService.Restore(clientUsername, data.objectIds)

          break
        }
        case "viewtrash": {
          resp = await rfsService.ViewTrash(clientUsername)

          break
        }
        case "rename": {
          const { success, error, data } = renameCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsService.Rename(
            clientUsername,
            data.parentDirectoryId,
            data.objectId,
            data.newName
          )

          break
        }
        case "move": {
          const { success, error, data } = moveCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsService.Move(
            clientUsername,
            data.fromParentDirectoryId,
            data.toParentDirectoryId,
            data.objectIds
          )

          break
        }
        case "copy": {
          const { success, error, data } = copyCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsService.Copy(
            clientUsername,
            data.fromParentDirectoryId,
            data.toParentDirectoryId,
            data.objectIds
          )

          break
        }
        default: {
          resp = `unknown command:${body.command}`
        }
      }

      return ws.send(helpers.WSReply(resp, body.command))
    } catch (error: any) {
      if (error.name === "AppError") {
        return ws.send(
          helpers.WSErrReply(error.code, error.error, body.command)
        )
      }

      console.error(error)
      return ws.send(
        helpers.WSErrReply(
          StatusCodes.INTERNAL_SERVER_ERROR,
          "internal server error",
          body.command
        )
      )
    }
  }
}
