import bcrypt from "bcrypt"
import jwt from "jsonwebtoken"

export function HashPassword(password: string) {
  return bcrypt.hash(password, 10)
}

export function PasswordMatchesHash(hash: string, plainPassword: string) {
  return bcrypt.compare(plainPassword, hash)
}

export function GetTokenAndExpiration() {
  let token: string
  const expires = Date.now() + 60 * 60 * 1000

  if (process.env.NODE_ENV != "production") {
    token = process.env.DUMMY_TOKEN || ""
  } else {
    token = String(Math.trunc(Math.random() * 899999 + 100000))
  }

  return { token, expires }
}

export function JwtSign(payload: object, secret: string, expiresIn: number) {
  return jwt.sign(payload, secret, { expiresIn: expiresIn })
}

export function JwtVerify(tokenString: string, secret: string) {
  return jwt.verify(tokenString, secret, {})
}
