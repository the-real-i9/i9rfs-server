import { StatusCodes } from "http-status-codes"
import * as rfsModel from "../../models/rfsModel.ts"
import type { DirT } from "../../appTypes.ts"
import appGlobals from "../../appGlobals.ts"
import * as user from "../../models/userModel.ts"

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

export async function deleteFilesInCS(
  clientUsername: string,
  fileCloudNames: string[]
) {
  console.time("Deleting files in cloud")

  let accFileSize = 0

  for (const fcn of fileCloudNames) {
    const file = appGlobals.AppGCSBucket.file(fcn)

    if (!(await file.exists())) {
      continue
    }

    const [metadata] = await file.getMetadata()

    accFileSize += Number(metadata.size || 0)

    await file.delete()
  }

  await user.UpdateStorageUsed(clientUsername, -accFileSize)

  console.timeEnd("Finished deleting files in cloud")
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
    await deleteFilesInCS(clientUsername, fileCloudNames)
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

export async function copyFilesInCS(
  clientUsername: string,
  fileCopyIdMaps: {
    cloud_object_name: string
    copy_id: string
  }[]
) {
  console.time("Copying files in cloud")

  let accFileSize = 0

  const year = new Date().getFullYear()
  const month = new Date().getMonth()

  for (const { cloud_object_name, copy_id } of fileCopyIdMaps) {
    const file = appGlobals.AppGCSBucket.file(cloud_object_name)

    if (!(await file.exists())) {
      continue
    }

    const [metadata] = await file.getMetadata()

    accFileSize += Number(metadata.size || 0)

    await file.copy(`uploads/${year}${month}/${copy_id}`)
  }

  await user.UpdateStorageUsed(clientUsername, accFileSize)

  console.timeEnd("Finished copying files in cloud")
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

  for (const oid of objectIds) {
    const { done, fileCopyIdMaps } = await rfsModel.Copy(
      clientUsername,
      fromParentDirectoryId,
      toParentDirectoryId,
      oid
    )

    if (done) {
      await copyFilesInCS(clientUsername, fileCopyIdMaps)
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
  const file = appGlobals.AppGCSBucket.file(data.cloudObjectName)

  const [uploaded] = await file.exists()
  if (!uploaded) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_FOUND,
      message: "object upload incomplete",
    }
  }

  const [metadata] = await file.getMetadata()

  const fileSize = Number(metadata.size || 0)
  if (!fileSize) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_ACCEPTABLE,
      message: "file has no content",
    }
  }

  const storageUsage = await user.StorageUsage(data.clientUsername)
  if (storageUsage.storage_used + fileSize >= storageUsage.alloc_storage) {
    await file.delete()
    throw {
      name: "AppError",
      code: StatusCodes.NOT_ACCEPTABLE,
      message:
        "uploaded file exceeds allocated storage space; file has been deleted",
    }
  }

  await user.UpdateStorageUsed(data.clientUsername, fileSize)

  const newFile = await rfsModel.Mkfil({
    ...data,
    mimeType: metadata.contentType || "",
    size: fileSize,
  })

  return newFile
}

export async function Download(data: {
  clientUsername: string
  objectId: string
  cloudObjectName: string
}) {
  const yes = await user.IsUserObject(data.clientUsername, data.objectId)
  if (!yes) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_FOUND,
      message: "file not found",
    }
  }

  const file = appGlobals.AppGCSBucket.file(data.cloudObjectName)

  const [exists] = await file.exists()
  if (!exists) {
    throw {
      name: "AppError",
      code: StatusCodes.NOT_FOUND,
      message: "file not found",
    }
  }

  const [downloadUrl] = await file.getSignedUrl({
    version: "v4",
    action: "read",
    expires: Date.now() + 1 * 24 * 60 * 60 * 1000, // 1 day: expected to be used immediately
  })

  return downloadUrl
}
