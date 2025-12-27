import { body, checkSchema } from "express-validator"
import { valResHandler } from "./valResHandler.ts"

export function RequestNewAccount() {
  return [
    checkSchema(
      {
        email: {
          isEmail: {
            errorMessage: "invalid email address",
          },
        },
      },
      ["body"]
    ),
    valResHandler,
  ]
}

export function VerifyEmail() {
  return [
    checkSchema(
      {
        code: {
          isString: true,
          isLength: {
            errorMessage: "invalid verification code",
            options: {
              min: 6,
              max: 6,
            },
          },
        },
      },
      ["body"]
    ),
    valResHandler,
  ]
}

export function RegisterUser() {
  return [
    checkSchema(
      {
        username: {
          isLength: {
            errorMessage: "username too short",
            options: {
              min: 3,
            },
          },
          matches: {
            errorMessage: "invalid username construct",
            options: /^[a-zA-Z0-9][a-zA-Z0-9_-]+[a-zA-Z0-9]$/,
          },
        },
        password: {
          isLength: {
            errorMessage: "minimum of 8 characters",
            options: {
              min: 8,
            },
          },
        },
      },
      ["body"]
    ),
    valResHandler,
  ]
}

export function Signin() {
  return [
    checkSchema(
      {
        emailOrUsername: {
          errorMessage: "invalid email or username",
          isEmail: {
            if: body("emailOrUsername").contains("@"),
          },
          isLength: {
            if: body("emailOrUsername").not().contains("@"),
            options: {
              min: 3,
            },
            bail: true,
          },
          matches: {
            if: body("emailOrUsername").not().contains("@"),
            options: /^[a-zA-Z0-9][a-zA-Z0-9_-]+[a-zA-Z0-9]$/,
          },
        },
        password: {
          isLength: {
            errorMessage: "minimum of 8 characters",
            options: {
              min: 8,
            },
          },
        },
      },
      ["body"]
    ),
    valResHandler,
  ]
}
