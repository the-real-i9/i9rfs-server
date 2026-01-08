import { checkSchema } from "express-validator"
import { valResHandler } from "./valResHandler.ts"

export function AuthorizeUpload() {
  return [
    checkSchema(
      {
        mimeType: {
          isMimeType: {
            errorMessage: "invalid MIME type",
          },
        },
        size: {
          isInt: {
            errorMessage: "expects size to be an integer greater than zero",
            options: {
              min: 1,
            },
          },
        },
      },
      ["body"]
    ),
    valResHandler,
  ]
}
