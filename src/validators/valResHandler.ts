import type { Request, Response, NextFunction } from "express"
import { validationResult } from "express-validator"
import { StatusCodes } from "http-status-codes"

export function valResHandler(req: Request, res: Response, next: NextFunction) {
  const result = validationResult(req)
  if (result.isEmpty()) {
    return next()
  }

  res.status(StatusCodes.BAD_REQUEST).send({ errors: result.array() })
}
