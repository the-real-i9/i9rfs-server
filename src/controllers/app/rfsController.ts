import type { IncomingMessage } from "node:http"
import type WebSocket from "ws"
import { StatusCodes } from "http-status-codes"
import { unpack } from "msgpackr"
import { GetSessionDataFromCookieHeader } from "../../helpers.ts"
import * as securityServices from "../../services/util/securityServices.ts"
import type { ClientUserT } from "../../appTypes.ts"
import {
  copyCommandValid,
  delCommandValid,
  downloadCommandValid,
  lsCommandValid,
  mkdirCommandValid,
  mkfilCommandValid,
  moveCommandValid,
  renameCommandValid,
  restoreCommandValid,
  rfsCommandBodyValid,
  trashCommandValid,
} from "./acValidation.ts"
import * as helpers from "../../helpers.ts"
import * as rfsService from "../../services/rfsService.ts"

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

    const { username: clientUsername } = authPayload as ClientUserT

    ws.on("message", async (data: Buffer<ArrayBufferLike>) => {
      const rfsCommandBody: { command: string; data: any } = unpack(data)

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
          case "mkfil": {
            const { success, error, data } = mkfilCommandValid(body.data)

            if (!success) {
              return ws.send(
                helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
              )
            }

            resp = await rfsService.Mkfil({ clientUsername, ...data })

            break
          }
          case "download": {
            const { success, error, data } = downloadCommandValid(body.data)

            if (!success) {
              return ws.send(
                helpers.WSErrReply(StatusCodes.BAD_REQUEST, error, body.command)
              )
            }

            resp = await rfsService.Download({ clientUsername, ...data })

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
    })
  } catch (error: any) {
    if (error.name === "TokenExpiredError") {
      return ws.close(1001, "401: session expired")
    }

    console.error(error)
    return ws.close(1001, "500: internal server error")
  }
}
