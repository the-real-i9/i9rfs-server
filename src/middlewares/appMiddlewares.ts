import type { Request, Response, NextFunction } from "express"
import { StatusCodes } from "http-status-codes"

export function UploadSessionCUC(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const ongoingUpload = req.session?.upload

    if (!ongoingUpload) {
      return res.status(StatusCodes.UNAUTHORIZED).json("no ongoing upload")
    }

    if (ongoingUpload.nextStep !== "cloud-upload-complete") {
      return res
        .status(StatusCodes.UNAUTHORIZED)
        .json("upload step out of turn")
    }

    if (ongoingUpload.cloudObjectName !== req.body.cloudObjectName) {
      return res
        .status(StatusCodes.BAD_REQUEST)
        .json("this is not the object being uploaded")
    }

    return next()
  } catch (error) {
    return next(error)
  }
}

export function UploadSessionCFO(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const ongoingUpload = req.session?.upload

    if (!ongoingUpload) {
      return res.status(StatusCodes.UNAUTHORIZED).json("no ongoing upload")
    }

    if (ongoingUpload.nextStep !== "create-file-object") {
      return res
        .status(StatusCodes.UNAUTHORIZED)
        .json("upload step out of turn")
    }

    if (ongoingUpload.cloudObjectName !== req.body.cloudObjectName) {
      return res
        .status(StatusCodes.BAD_REQUEST)
        .json("this is not the object being uploaded")
    }

    return next()
  } catch (error) {
    return next(error)
  }
}
