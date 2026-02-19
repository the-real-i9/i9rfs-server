import { StatusCodes } from "http-status-codes"
import * as rfsModel from "../models/rfsModel.ts"
import type { DirT } from "../appTypes.ts"
import * as userService from "./userService.ts"
import * as cloudStorageService from "./cloudStorageService.ts"

export function Ls(clientUsername: string, directoryId: string) {
  return rfsModel.Ls(clientUsername, directoryId)
}

export async function Mkdir(
  clientUsername: string,
  parentDirectoryId: string,
  directoryNames: string[]
) {
  const newDirs: (DirT | null)[] = []

  for (const dirName of directoryNames) {
    if (!dirName.includes("/")) {
      newDirs.push(
        await rfsModel.Mkdir(clientUsername, parentDirectoryId, dirName)
      )
    } else {
      const subDirs = dirName.split("/")

      let outerDirId: string = parentDirectoryId

      for (let i = 0; i < subDirs.length; i++) {
        const innerDir = await rfsModel.Mkdir(
          clientUsername,
          outerDirId,
          subDirs[i] || ""
        )

        outerDirId = innerDir?.id || ""

        if (i === 0) {
          newDirs.push(innerDir)
        }
      }
    }
  }

  return newDirs
}

export async function Del(
  clientUsername: string,
  parentDirectoryId: string,
  objectIds: string[]
) {
  const { done, fileCloudNames } = await rfsModel.Del(
    clientUsername,
    parentDirectoryId,
    objectIds
  )

  if (done) {
    await cloudStorageService.DeleteFilesInCS(
      fileCloudNames,
      async (deletedFilesSize: number) => {
        await userService.UpdateStorageUsed(clientUsername, -deletedFilesSize)
      }
    )
  }

  return done
}

export function Trash(
  clientUsername: string,
  parentDirectoryId: string,
  objectIds: string[]
) {
  return rfsModel.Trash(clientUsername, parentDirectoryId, objectIds)
}

export function Restore(clientUsername: string, objectIds: string[]) {
  return rfsModel.Restore(clientUsername, objectIds)
}

export function ViewTrash(clientUsername: string) {
  return rfsModel.ViewTrash(clientUsername)
}

export function Rename(
  clientUsername: string,
  parentDirectoryId: string,
  objectId: string,
  newName: string
) {
  return rfsModel.Rename(clientUsername, parentDirectoryId, objectId, newName)
}

export function Move(
  clientUsername: string,
  fromParentDirectoryId: string,
  toParentDirectoryId: string,
  objectIds: string[]
) {
  if (fromParentDirectoryId === toParentDirectoryId) {
    throw {
      name: "AppError",
      code: StatusCodes.BAD_REQUEST,
      message: "attempt to move to the same directory",
    }
  }

  return rfsModel.Move(
    clientUsername,
    fromParentDirectoryId,
    toParentDirectoryId,
    objectIds
  )
}

export async function Copy(
  clientUsername: string,
  fromParentDirectoryId: string,
  toParentDirectoryId: string,
  objectIds: string[]
) {
  if (fromParentDirectoryId === toParentDirectoryId) {
    throw {
      name: "AppError",
      code: StatusCodes.BAD_REQUEST,
      message: "attempt to copy to the same directory",
    }
  }

  const now = Date.now()

  for (const oid of objectIds) {
    const { done, fileCopyMaps } = await rfsModel.Copy(
      clientUsername,
      fromParentDirectoryId,
      toParentDirectoryId,
      oid,
      now
    )

    if (done) {
      await cloudStorageService.CopyFilesInCS(
        now,
        fileCopyMaps,
        async (copiedFilesSize: number) => {
          await userService.UpdateStorageUsed(clientUsername, copiedFilesSize)
        }
      )
    }
  }

  return true
}

export async function Mkfil(data: {
  clientUsername: string
  parentDirectoryId: string
  objectId: string
  cloudObjectName: string
  filename: string
}) {
  const {
    exists,
    size: fileSize,
    contentType: mimeType,
  } = await cloudStorageService.FileExistsInCS(data.cloudObjectName)

  if (!exists) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_FOUND,
      message: "file not found in cloud",
    }
  }

  if (!fileSize) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_ACCEPTABLE,
      message: "file has no content",
    }
  }

  const storageUsage = await userService.GetStorageUsage(data.clientUsername)
  if (storageUsage.storage_used + fileSize >= storageUsage.alloc_storage) {
    await cloudStorageService.DeleteExistingFileInCS(data.cloudObjectName)
    throw {
      name: "AppError",
      code: StatusCodes.NOT_ACCEPTABLE,
      message:
        "uploaded file exceeds allocated storage space; file has been deleted",
    }
  }

  await userService.UpdateStorageUsed(data.clientUsername, fileSize)

  const newFile = await rfsModel.Mkfil({
    ...data,
    mimeType,
    size: fileSize,
  })

  return newFile
}

export async function Download(data: {
  clientUsername: string
  fileObjectId: string
}) {
  const fileCloudObjectName = await rfsModel.Download(
    data.clientUsername,
    data.fileObjectId
  )
  if (!fileCloudObjectName) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_FOUND,
      message: "file not found",
    }
  }

  const { exists } = await cloudStorageService.FileExistsInCS(
    fileCloudObjectName
  )

  if (!exists) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_FOUND,
      message: "file not found",
    }
  }

  const downloadUrl = await cloudStorageService.GetDownloadUrl(
    fileCloudObjectName
  )

  return downloadUrl
}
