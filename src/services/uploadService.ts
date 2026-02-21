import { v4 as uuid } from "uuid"
import * as userService from "./userService.ts"
import * as cloudStorageService from "./cloudStorageService.ts"
import { StatusCodes } from "http-status-codes"

export async function AuthorizeUpload(
  username: string,
  mimeType: string,
  size: number
) {
  const storageUsage = await userService.GetStorageUsage(username)
  if (storageUsage.storage_used + size >= storageUsage.alloc_storage) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_ACCEPTABLE,
      message: "not enough space!",
    }
  }

  const objectId = uuid()
  const cloudObjectName = `uploads/${new Date().getFullYear()}${new Date().getMonth()}/${objectId}`

  const url = await cloudStorageService.GetUploadUrl(cloudObjectName, mimeType)

  return { uploadUrl: url, objectId, cloudObjectName }
}
