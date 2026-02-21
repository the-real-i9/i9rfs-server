import appGlobals from "../appGlobals.ts"

export async function GetUploadUrl(
  cloudObjectName: string,
  contentType: string
) {
  const [url] = await appGlobals.AppGCSBucket.file(
    cloudObjectName
  ).getSignedUrl({
    version: "v4",
    action: "resumable",
    expires: Date.now() + 15 * 60 * 1000, // 15 minutes
    contentType,
  })

  return url
}

export async function GetDownloadUrl(cloudObjectName: string) {
  const file = appGlobals.AppGCSBucket.file(cloudObjectName)

  const [url] = await file.getSignedUrl({
    version: "v4",
    action: "read",
    expires: Date.now() + 1 * 24 * 60 * 60 * 1000, // 1 day: expected to be used immediately
  })

  return url
}

export async function FileExistsInCS(cloudObjectName: string) {
  const file = appGlobals.AppGCSBucket.file(cloudObjectName)

  let size = 0
  let contentType = ""

  const [exists] = await file.exists()
  if (exists) {
    const [metadata] = await file.getMetadata()

    size = Number(metadata.size) || 0
    contentType = metadata.contentType || ""
  }

  return { exists, size, contentType }
}

export async function DeleteExistingFileInCS(cloudObjectName: string) {
  const file = appGlobals.AppGCSBucket.file(cloudObjectName)

  await file.delete()
}

export async function DeleteFilesInCS(
  fileCloudNames: string[],
  callback: (deletedFilesSize: number) => Promise<void>
) {
  if (!fileCloudNames.length) {
    return
  }

  console.log("Deleting files in cloud")

  let deletedFilesSize = 0

  for (const fcn of fileCloudNames) {
    const file = appGlobals.AppGCSBucket.file(fcn)

    if (!(await file.exists())) {
      continue
    }

    const [metadata] = await file.getMetadata()

    deletedFilesSize += Number(metadata.size || 0)

    await file.delete()
  }

  await callback(deletedFilesSize)

  console.log("Finished deleting files in cloud")
}

export async function CopyFilesInCS(
  now: number,
  fileCopyMaps: {
    cloud_object_name: string
    copy_id: string
  }[],
  callback: (copiedFilesSize: number) => Promise<void>
) {
  if (!fileCopyMaps.length) {
    return
  }

  console.log("Copying files in cloud")

  let copiedFilesSize = 0

  const year = new Date(now).getFullYear()
  const month = new Date(now).getMonth()

  for (const { cloud_object_name, copy_id } of fileCopyMaps) {
    const file = appGlobals.AppGCSBucket.file(cloud_object_name)

    if (!(await file.exists())) {
      continue
    }

    const [metadata] = await file.getMetadata()

    copiedFilesSize += Number(metadata.size || 0)

    await file.copy(`uploads/${year}${month}/${copy_id}`)
  }

  await callback(copiedFilesSize)

  console.log("Finished copying files in cloud")
}
