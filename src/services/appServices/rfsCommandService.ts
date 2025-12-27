import { StatusCodes } from "http-status-codes"
import * as rfsCommandModel from "../../models/rfsCommandModel.ts"
import type { DirT } from "../../appTypes.ts"

export function Ls(clientUsername: string, directoryId: string) {
  return rfsCommandModel.Ls(clientUsername, directoryId)
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
        await rfsCommandModel.Mkdir(clientUsername, parentDirectoryId, dirName)
      )
    } else {
      const subDirs = dirName.split("/")

      let outerDirId: string = parentDirectoryId

      for (let i = 0; i < subDirs.length; i++) {
        const innerDir = await rfsCommandModel.Mkdir(
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
  const { done, fileIds } = await rfsCommandModel.Del(
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
  return rfsCommandModel.Trash(clientUsername, parentDirectoryId, objectIds)
}

export function Restore(clientUsername: string, objectIds: string[]) {
  return rfsCommandModel.Restore(clientUsername, objectIds)
}

export function ViewTrash(clientUsername: string) {
  return rfsCommandModel.ViewTrash(clientUsername)
}

export function Rename(
  clientUsername: string,
  parentDirectoryId: string,
  objectId: string,
  newName: string
) {
  return rfsCommandModel.Rename(
    clientUsername,
    parentDirectoryId,
    objectId,
    newName
  )
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

  return rfsCommandModel.Move(
    clientUsername,
    fromParentDirectoryId,
    toParentDirectoryId,
    objectIds
  )
}

export function copyFilesInCS(fileCopyIdMaps: any[]) {}

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
    const { done, fileCopyIdMaps } = await rfsCommandModel.Copy(
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
