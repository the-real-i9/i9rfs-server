import type { Request, Response } from "express"
import type { IncomingMessage } from "node:http"
import type WebSocket from "ws"

import type { ClientUserT } from "../../appTypes.ts"
import { GetSessionDataFromCookieHeader } from "../../helpers.ts"
import * as securityServices from "../../services/appServices/securityServices.ts"
import type { JwtPayload } from "jsonwebtoken"
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
import * as rfsCommandService from "../../services/appServices/rfsCommandService.ts"

export async function Signout(req: Request, res: Response) {
  const clientUser: ClientUserT = res.locals.user

  req.session = null

  return res.json(`Bye, ${clientUser.username}! See you again!`)
}

export function RFSController(ws: WebSocket, request: IncomingMessage) {
  try {
    const sessionData = GetSessionDataFromCookieHeader(
      request.headers.cookie || ""
    )
    if (!sessionData || !sessionData.user) {
      return ws.close(1000, "401: authentication required")
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
      return ws.close(1000, "401: session expired")
    }

    console.error(error)
    return ws.close(1001, "500: internal server error")
  }
}

function wsMessageHandler(ws: WebSocket, clientUsername: string) {
  return async function (data: WebSocket.RawData) {
    const rfsCommandBody = JSON.parse(data.toString())

    const { success, error, data: body } = rfsCommandBodyValid(rfsCommandBody)
    if (!success) {
      return ws.send(
        helpers.WSErrReply(
          StatusCodes.BAD_REQUEST,
          error,
          rfsCommandBody.command
        )
      )
    }

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

          resp = await rfsCommandService.Ls(clientUsername, data.directoryId)
        }
        case "mkdir": {
          const { success, error, data } = mkdirCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsCommandService.Mkdir(
            clientUsername,
            data.parentDirectoryId,
            data.directoryName
          )
        }
        case "del": {
          const { success, error, data } = delCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsCommandService.Del(
            clientUsername,
            data.parentDirectoryId,
            data.objectIds
          )
        }
        case "trash": {
          const { success, error, data } = trashCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsCommandService.Trash(
            clientUsername,
            data.parentDirectoryId,
            data.objectIds
          )
        }
        case "restore": {
          const { success, error, data } = restoreCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsCommandService.Restore(clientUsername, data.objectIds)
        }
        case "vwtrash": {
          resp = await rfsCommandService.ViewTrash(clientUsername)
        }
        case "rename": {
          const { success, error, data } = renameCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsCommandService.Rename(
            clientUsername,
            data.parentDirectoryId,
            data.objectId,
            data.newName
          )
        }
        case "move": {
          const { success, error, data } = moveCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsCommandService.Move(
            clientUsername,
            data.fromParentDirectoryId,
            data.toParentDirectoryId,
            data.objectIds
          )
        }
        case "copy": {
          const { success, error, data } = copyCommandValid(body.data)

          if (!success) {
            return ws.send(
              helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
            )
          }

          resp = await rfsCommandService.Copy(
            clientUsername,
            data.fromParentDirectoryId,
            data.toParentDirectoryId,
            data.objectIds
          )
        }
        case "upload": {
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
