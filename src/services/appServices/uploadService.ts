import { v4 as uuid } from "uuid"
import appGlobals from "../../appGlobals.ts"
import * as user from "../../models/userModel.ts"
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

export async function CloudUploadComplete(
  username: string,
  cloudObjectName: string
) {
  const file = appGlobals.AppGCSBucket.file(cloudObjectName)

  const [uploaded] = await file.exists()
  if (!uploaded) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_FOUND,
      message: "cloud upload incomplete",
    }
  }

  const [metadata] = await file.getMetadata()

  const fileSize = Number(metadata.size)
  if (!fileSize) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_ACCEPTABLE,
      message: "file has no content",
    }
  }

  const storageUsage = await user.StorageUsage(username)
  if (storageUsage.storage_used + fileSize >= storageUsage.alloc_storage) {
    await file.delete()
    throw {
      name: "AppError",
      code: StatusCodes.NOT_ACCEPTABLE,
      message:
        "uploaded file exceeds allocated storage space; file has been deleted",
    }
  }

  await user.UpdateStorageUsed(username, fileSize)
}
