import type { IncomingMessage } from "node:http"
import { GetSessionDataFromCookieHeader } from "../../helpers.ts"
import * as securityServices from "../../services/util/securityServices.ts"
import type { ClientUserT } from "../../appTypes.ts"
import type WebSocket from "ws"
import { WSMessageHandler } from "./appControllers.ts"

export function RealtimeController(ws: WebSocket, request: IncomingMessage) {
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

    ws.on("message", WSMessageHandler(ws, clientUser))
  } catch (error: any) {
    if (error.name === "TokenExpiredError") {
      return ws.close(1001, "401: session expired")
    }

    console.error(error)
    return ws.close(1001, "500: internal server error")
  }
}
