import * as user from "../models/userModel.ts"

export function CreateNewUser(
  email: string,
  username: string,
  password: string
) {
  return user.New(email, username, password)
}

export function SigninFindUser(emailOrUsername: string) {
  return user.SigninFind(emailOrUsername)
}

export function UserExists(emailOrUsername: string) {
  return user.Exists(emailOrUsername)
}

export function UpdateStorageUsed(username: string, delta: number) {
  return user.UpdateStorageUsed(username, delta)
}

export function GetStorageUsage(username: string) {
  return user.StorageUsage(username)
}
