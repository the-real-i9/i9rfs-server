export type ClientUserT = {
  username: string
}

export type DirT = {
  date_created: string
  date_modified: string
  starred: boolean
  native: boolean
  name: string
  obj_type: "directory" | "file"
  id: string
}
