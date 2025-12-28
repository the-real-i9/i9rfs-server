export type ClientUserT = {
  username: string
}

export type DirT = {
  id: string
  obj_type: "directory" | "file"
  name: string
  native: boolean
  starred: boolean
  date_created: string
  date_modified: string
  trashed_on: string
}
