import fs from "node:fs/promises"

import { type TestContext } from "node:test"
import { type DirT } from "../src/appTypes.ts"

export function containsDirs(
  actual: DirT[],
  expectedDirs: string[],
  t: TestContext
) {
  for (const dirName of expectedDirs) {
    const dir = actual.find((d) => d.name === dirName)
    t.assert.ok(dir)
    t.assert.ok(dir.id)
    t.assert.strictEqual(dir.obj_type, "directory")
  }
}

export function notContainsDirs(
  actual: DirT[],
  notExpectedDirs: string[],
  t: TestContext
) {
  const actualDirs = actual.map((d) => d.name)

  for (const dirName of notExpectedDirs) {
    t.assert.ok(!actualDirs.includes(dirName))
  }
}

export async function startResumableUpload(
  uploadUrl: string,
  contentType: string
) {
  const res = await fetch(uploadUrl, {
    method: "POST",
    headers: {
      "Content-Type": contentType,
      "x-goog-resumable": "start",
    },
  })

  if (!res.ok) {
    throw new Error(`Failed to start resumable upload: ${res.status}`)
  }

  const sessionUrl = res.headers.get("location")
  if (!sessionUrl) {
    throw new Error("No resumable session URL returned")
  }

  return sessionUrl
}

const CHUNK_SIZE = 256 * 1024 // 256 KeB

export async function uploadFileInChunks(
  sessionUrl: string,
  filePath: string,
  contentType: string,
  onProgress: (offset: number, fileSize: number) => void
) {
  const stat = await fs.stat(filePath)
  const fileSize = stat.size

  const fd = await fs.open(filePath, "r")

  let offset = 0
  const buffer = Buffer.alloc(CHUNK_SIZE)

  try {
    while (offset < fileSize) {
      const bytesToRead = Math.min(CHUNK_SIZE, fileSize - offset)
      const { bytesRead } = await fd.read(buffer, 0, bytesToRead, offset)

      const chunk = buffer.subarray(0, bytesRead)
      const end = offset + bytesRead - 1

      const res = await fetch(sessionUrl, {
        method: "PUT",
        headers: {
          "Content-Type": contentType,
          "Content-Length": bytesRead.toString(),
          "Content-Range": `bytes ${offset}-${end}/${fileSize}`,
        },
        body: chunk,
      })

      if (res.status === 308) {
        offset += bytesRead
        onProgress(offset, fileSize)
        continue
      }

      if (res.ok) {
        offset += bytesRead
        onProgress(fileSize, fileSize)
        return
      }

      throw new Error(`Chunk upload failed: ${res.status}`)
    }
  } finally {
    await fd.close()
  }
}

export function logProgress(sent: number, total: number) {
  const percent = ((sent / total) * 100).toFixed(2)
  console.log(`Upload progress: ${percent}% (${sent}/${total})`)
}
