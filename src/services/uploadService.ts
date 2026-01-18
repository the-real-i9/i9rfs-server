import { v4 as uuid } from "uuid"
import appGlobals from "../appGlobals.ts"
import * as user from "../models/userModel.ts"
import { StatusCodes } from "http-status-codes"

export async function AuthorizeUpload(
  username: string,
  mimeType: string,
  size: number
) {
  const storageUsage = await user.StorageUsage(username)
  if (storageUsage.storage_used + size >= storageUsage.alloc_storage) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_ACCEPTABLE,
      message: "not enough space!",
    }
  }

  const objectId = uuid()
  const cloudObjectName = `uploads/${new Date().getFullYear()}${new Date().getMonth()}/${objectId}`

  const [url] = await appGlobals.AppGCSBucket.file(
    cloudObjectName
  ).getSignedUrl({
    version: "v4",
    action: "resumable",
    expires: Date.now() + 15 * 60 * 1000, // 15 minutes
    contentType: mimeType,
  })

  return { uploadUrl: url, objectId, cloudObjectName }
}
