export type ClientUserT = {
  username: string
}

export type DirT = {
  id: string
  obj_type: "directory"
  name: string
  native: boolean
  starred: boolean
  date_created: number
  date_modified: number
  trashed_on: number
}

export type FileT = {
  id: string
  obj_type: "file"
  name: string
  cloud_object_name: string
  mime_type: string
  size: number
  starred: boolean
  date_created: number
  date_modified: number
  trashed_on: number
}
