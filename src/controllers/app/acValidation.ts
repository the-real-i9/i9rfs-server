import * as z from "zod"

export function rfsCommandBodyValid(body: any) {
  const schema = z.object({
    command: z.string(),
    data: body.command === "viewtrash" ? z.any().optional() : z.any(),
  })

  const res = schema.safeParse(body)

  return res
}

export function lsCommandValid(command: any) {
  const schema = z.object({
    directoryId: z.literal("/").or(z.uuid()),
  })

  const res = schema.safeParse(command)

  return res
}

export function mkdirCommandValid(command: any) {
  const schema = z.object({
    parentDirectoryId: z.literal("/").or(z.uuid()),
    directoryNames: z.array(z.string()),
  })

  const res = schema.safeParse(command)

  return res
}

export function delCommandValid(command: any) {
  const schema = z.object({
    parentDirectoryId: z.literal("/").or(z.uuid()),
    objectIds: z.array(z.uuid()),
  })

  const res = schema.safeParse(command)

  return res
}

export function trashCommandValid(command: any) {
  const schema = z.object({
    parentDirectoryId: z.literal("/").or(z.uuid()),
    objectIds: z.array(z.uuid()),
  })

  const res = schema.safeParse(command)

  return res
}

export function restoreCommandValid(command: any) {
  const schema = z.object({
    objectIds: z.array(z.uuid()),
  })

  const res = schema.safeParse(command)

  return res
}

export function renameCommandValid(command: any) {
  const schema = z.object({
    parentDirectoryId: z.literal("/").or(z.uuid()),
    objectId: z.uuid(),
    newName: z.string(),
  })

  const res = schema.safeParse(command)

  return res
}

export function moveCommandValid(command: any) {
  const schema = z.object({
    fromParentDirectoryId: z.literal("/").or(z.uuid()),
    toParentDirectoryId: z.literal("/").or(z.uuid()),
    objectIds: z.array(z.uuid()),
  })

  const res = schema.safeParse(command)

  return res
}

export function copyCommandValid(command: any) {
  const schema = z.object({
    fromParentDirectoryId: z.literal("/").or(z.uuid()),
    toParentDirectoryId: z.literal("/").or(z.uuid()),
    objectIds: z.array(z.uuid()),
  })

  const res = schema.safeParse(command)

  return res
}

export function mkfilCommandValid(command: any) {
  const schema = z.object({
    parentDirectoryId: z.literal("/").or(z.uuid()),
    objectId: z.uuid(),
    cloudObjectName: z
      .string()
      .regex(/^uploads\/[\w-/]+\w$/, { error: "invalid cloud object name" }),
    filename: z.string().regex(/^.+\.\w+$/, { error: "invalid file name" }),
  })

  const res = schema.safeParse(command)

  return res
}

export function downloadCommandValid(command: any) {
  const schema = z.object({
    fileObjectId: z.uuid(),
  })

  const res = schema.safeParse(command)

  return res
}
