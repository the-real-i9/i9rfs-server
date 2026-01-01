import type { ClientUserT } from "../appTypes.ts"
import * as db from "./db/db.ts"

export async function New(email: string, username: string, password: string) {
  const res = await db.WriteQuery(
    `/* cypher */
		CREATE (u:User { email: $email, username: $username, password: $password, storage_used: 0, alloc_storage: $alloc_storage })
		
		CREATE (root:UserRoot{ user: $username }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Documents", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Downloads", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Music", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Pictures", date_created: $now, date_modified: $now, native: true, starred: false }),
			(root)-[:HAS_CHILD]->(:Object{ id: randomUUID(), obj_type: "directory", name: "Videos", date_created: $now, date_modified: $now, native: true, starred: false })
		
		CREATE (:UserTrash{ user: $username })
			
		RETURN u { .username } AS new_user
		`,
    {
      email,
      username,
      password,
      alloc_storage: 50 * 1024 ** 3,
      now: Date.now(),
    }
  )

  return res.records[0]?.get("new_user") as ClientUserT
}

export async function AuthFind(emailOrUsername: string) {
  const res = await db.ReadQuery(
    `/* cypher */
		MATCH (u:User)
		WHERE u.email = $emailOrUsername OR u.username = $emailOrUsername
		RETURN u { .username, .password } AS found_user
		`,
    {
      emailOrUsername,
    }
  )

  if (res.records.length === 0) {
    return null
  }

  return res.records[0]?.get("found_user") as {
    username: string
    password: string
  }
}

export async function Exists(emailOrUsername: string) {
  const res = await db.ReadQuery(
    `/* cypher */
		RETURN EXISTS {
			MATCH (u:User)
			WHERE u.email = $emailOrUsername OR u.username = $emailOrUsername
		} AS user_exists
		`,
    {
      emailOrUsername,
    }
  )

  return res.records[0]?.get("user_exists") as boolean
}

export async function StorageUsage(username: string) {
  const res = await db.ReadQuery(
    `/* cypher */
		  MATCH (u:User { username: $username })
		  RETURN u { .storage_used, .alloc_storage } as storage_usage
		`,
    {
      username,
    }
  )

  return res.records[0]?.get("storage_usage") as {
    storage_used: number
    alloc_storage: number
  }
}

export async function UpdateStorageUsed(username: string, delta: number) {
  const res = await db.WriteQuery(
    `/* cypher */
		  MATCH (u:User { username: $username })
      SET u.storage_used = u.storage_used + $delta
		  RETURN true AS done
		`,
    {
      username,
      delta,
    }
  )

  return res.records[0]?.get("done") as boolean
}
