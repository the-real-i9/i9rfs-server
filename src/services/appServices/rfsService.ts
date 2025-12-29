import { StatusCodes } from "http-status-codes"
import * as rfsModel from "../../models/rfsModel.ts"
import type { DirT } from "../../appTypes.ts"
import appGlobals from "../../appGlobals.ts"

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

export function deleteFilesInCS(fileIds: any[]) {}

export async function Del(
  clientUsername: string,
  parentDirectoryId: string,
  objectIds: string[]
) {
  const { done, fileIds } = await rfsModel.Del(
    clientUsername,
    parentDirectoryId,
    objectIds
  )

  if (done) {
    deleteFilesInCS(fileIds)
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

export function copyFilesInCS(
  fileCopyIdMaps: {
    copied_id: string
    copy_id: string
  }[]
) {}

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
      copyFilesInCS(fileCopyIdMaps)
    }
  }

  return true
}

export async function CreateFile(
  clientUsername: string,
  parentDirectoryId: string,
  objectId: string,
  cloudObjectName: string,
  displayName: string
) {
  const file = appGlobals.AppGCSBucket.file(cloudObjectName)

  const [metadata] = await file.getMetadata()

  const done = await rfsModel.CreateFile(
    clientUsername,
    parentDirectoryId,
    objectId,
    cloudObjectName,
    displayName,
    metadata.contentType || "",
    Number(metadata.size)
  )

  return done
}
