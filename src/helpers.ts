import { pack } from "msgpackr"

export function GetSessionDataFromCookieHeader(cookieHeader: string) {
  const cookies = Object.fromEntries(
    cookieHeader.split(";").map((c) => c.trim().split("="))
  )

  const rawSession = cookies.session

  if (!rawSession) return null

  const [payloadB64, signature] = rawSession.split(".")
  const json = Buffer.from(payloadB64, "base64").toString("utf8")
  const sessionObj = JSON.parse(json)

  return sessionObj
}

export function WSErrReply(errCode: number, err: any, toCommand: string) {
  return pack({
    event: "server error",
    toCommand,
    data: {
      statusCode: errCode,
      error: err,
    },
  })
}

export function WSReply(data: any, toCommand: string) {
  return pack({
    event: "server reply",
    toCommand,
    data,
  })
}
