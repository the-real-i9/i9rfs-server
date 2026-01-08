import appGlobals from "../appGlobals.ts"
import type { DirT, FileT } from "../appTypes.ts"
import * as db from "./db/db.ts"

export async function Ls(clientUsername: string, directoryId: string) {
  let matchFromPath: string

  if (directoryId === "/") {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.trashed IS NULL)"
  } else {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $directory_id })-[:HAS_CHILD]->(obj WHERE obj.trashed IS NULL)"
  }

  const res = await db.ReadQuery(
    `/* cypher */
			MATCH ${matchFromPath}
			WITH obj, obj.date_created AS date_created, obj.date_modified AS date_modified
			ORDER BY obj.obj_type DESC, obj.name ASC
			RETURN collect(obj { .*, date_created, date_modified }) AS dir_cont
		`,
    {
      client_username: clientUsername,
      directory_id: directoryId,
    }
  )

  if (!res.records.length) {
    return null
  }

  return res.records[0]?.get("dir_cont") as (FileT | DirT)[]
}

export async function Mkdir(
  clientUsername: string,
  parentDirectoryId: string,
  directoryName: string
) {
  let matchFromPath: string
  let matchFromIdent: string

  if (parentDirectoryId === "/") {
    matchFromPath = "/* cypher */(root:UserRoot{ user: $client_username })"
    matchFromIdent = "/* cypher */root"
  } else {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(parObj:Object{ id: $parent_dir_id })"
    matchFromIdent = "/* cypher */parObj"
  }

  const res = await db.WriteQuery(
    `/* cypher */
			MATCH ${matchFromPath}
			CREATE (${matchFromIdent})-[:HAS_CHILD]->(obj:Object{ id: randomUUID(), obj_type: "directory", name: $dir_name, date_created: $now, date_modified: $now })
			
			RETURN obj { .* } AS new_dir
		`,
    {
      client_username: clientUsername,
      parent_dir_id: parentDirectoryId,
      dir_name: directoryName,
      now: Date.now(),
    }
  )

  if (!res.records.length) {
    return null
  }

  return res.records[0]?.get("new_dir") as DirT
}

export async function Del(
  clientUsername: string,
  parentDirectoryId: string,
  objectIds: string[]
) {
  let matchFromPath: string

  if (parentDirectoryId === "/") {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)(()-[:HAS_CHILD]->(childObjs))*"
  } else {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)(()-[:HAS_CHILD]->(childObjs))*"
  }

  const res = await db.WriteQuery(
    `/* cypher */
			MATCH ${matchFromPath}
			
			WITH obj, childObjs,
				[o IN collect(obj) WHERE o.obj_type = "file" | o.cloud_object_name] AS objFileCloudNames,
				[co IN childObjs WHERE co.obj_type = "file" | co.cloud_object_name] AS childObjFileCloudNames

			DETACH DELETE obj

			WITH objFileCloudNames, childObjFileCloudNames, childObjs

			UNWIND (childObjs + [null]) AS cObj
			DETACH DELETE cObj

			WITH objFileCloudNames, childObjFileCloudNames

			RETURN objFileCloudNames + childObjFileCloudNames AS file_cloud_names
		`,
    {
      client_username: clientUsername,
      parent_dir_id: parentDirectoryId,
      object_ids: objectIds,
    }
  )

  if (!res.records.length) {
    return { done: false, fileCloudNames: [] }
  }

  return {
    done: true,
    fileCloudNames: res.records[0]?.get("file_cloud_names") as string[],
  }
}

export async function Trash(
  clientUsername: string,
  parentDirectoryId: string,
  objectIds: string[]
) {
  let matchFromPath: string

  if (parentDirectoryId === "/") {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL)"
  } else {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)"
  }

  const res = await db.WriteQuery(
    `/* cypher */
			MATCH ${matchFromPath}

			SET obj.trashed = true, obj.trashed_on = $now

			WITH obj

			MATCH (trash:UserTrash{ user: $client_username })
			CREATE (trash)-[:HAS_CHILD]->(obj)

			RETURN true AS done
		`,
    {
      client_username: clientUsername,
      parent_dir_id: parentDirectoryId,
      object_ids: objectIds,
      now: Date.now(),
    }
  )

  if (!res.records.length) {
    return false
  }

  return true
}

export async function Restore(clientUsername: string, objectIds: string[]) {
  const res = await db.WriteQuery(
    `/* cypher */
		MATCH (:UserTrash{ user: $client_username })-[tr:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)

		DELETE tr

		SET obj.trashed = null, obj.trashed_on = null

		RETURN true AS done
		`,
    {
      client_username: clientUsername,
      object_ids: objectIds,
    }
  )
  if (!res.records.length) {
    return false
  }

  return true
}

export async function ViewTrash(clientUsername: string) {
  const res = await db.ReadQuery(
    `/* cypher */
		MATCH (:UserTrash{ user: $client_username })-[:HAS_CHILD]->(obj)

		ORDER BY obj.obj_type DESC, obj.name ASC
		RETURN collect(obj { .* }) AS trash_cont
		`,
    {
      client_username: clientUsername,
    }
  )

  if (!res.records.length) {
    return []
  }

  return res.records[0]?.get("trash_cont") as DirT[]
}

export async function Rename(
  clientUsername: string,
  parentDirectoryId: string,
  objectId: string,
  newName: string
) {
  let matchFromPath: string

  if (parentDirectoryId === "/") {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->(obj:Object{ id: $object_id } WHERE obj.native IS NULL AND obj.trashed IS NULL)"
  } else {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(:Object{ id: $parent_dir_id })-[:HAS_CHILD]->(obj:Object{ id: $object_id } WHERE obj.trashed IS NULL)"
  }

  const res = await db.WriteQuery(
    `/* cypher */
			MATCH ${matchFromPath}

			SET obj.name = $new_name, obj.date_modified = $now

			RETURN true AS done
		`,
    {
      client_username: clientUsername,
      parent_dir_id: parentDirectoryId,
      object_id: objectId,
      new_name: newName,
      now: Date.now(),
    }
  )

  if (!res.records.length) {
    return false
  }

  return true
}

export async function Move(
  clientUsername: string,
  fromParentDirectoryId: string,
  toParentDirectoryId: string,
  objectIds: string[]
) {
  let cypher: string

  if (fromParentDirectoryId === "/" && toParentDirectoryId !== "/") {
    cypher = `/* cypher */
			MATCH (root:UserRoot{ user: $client_username }),
				(root)-[old:HAS_CHILD]->(obj WHERE obj.id IN $object_ids AND obj.native IS NULL),
				(root)-[:HAS_CHILD]->+(toParDir:Object{ id: $to_parent_dir_id })
				
			CREATE (toParDir)-[:HAS_CHILD]->(obj)
			SET toParDir.date_modified = $now

			DELETE old

			RETURN true AS done
		`
  } else if (fromParentDirectoryId !== "/" && toParentDirectoryId === "/") {
    cypher = `/* cypher */
			MATCH (root:UserRoot{ user: $client_username }),
				(root)-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })-[old:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)

			CREATE (root)-[:HAS_CHILD]->(obj)
			SET fromParDir.date_modified = $now

			DELETE old

			RETURN true AS done
		`
  } else {
    cypher = `/* cypher */
			MATCH (root:UserRoot{ user: $client_username }),
				(root)-[:HAS_CHILD]->+(toParDir:Object{ id: $to_parent_dir_id }),
				(root)-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })-[old:HAS_CHILD]->(obj WHERE obj.id IN $object_ids)

			CREATE (toParDir)-[:HAS_CHILD]->(obj)
			SET fromParDir.date_modified = $now, toParDir.date_modified = $now

			DELETE old

			RETURN true AS done
		`
  }

  const res = await db.WriteQuery(cypher, {
    client_username: clientUsername,
    from_parent_dir_id: fromParentDirectoryId,
    to_parent_dir_id: toParentDirectoryId,
    object_ids: objectIds,
    now: Date.now(),
  })
  if (!res.records.length) {
    return false
  }

  return true
}

export async function Copy(
  clientUsername: string,
  fromParentDirectoryId: string,
  toParentDirectoryId: string,
  objectId: string
) {
  const sess = appGlobals.Neo4jDriver.session()

  const res = await sess.executeWrite(async (tx) => {
    let matchFromPath: string, matchFromIdent: string
    let matchToPath: string, matchToIdent: string

    const now = Date.now()

    if (fromParentDirectoryId === "/") {
      matchFromPath = "/* cypher */(root:UserRoot{ user: $client_username })"
      matchFromIdent = "/* cypher */root"
    } else {
      matchFromPath =
        "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(fromParDir:Object{ id: $from_parent_dir_id })"
      matchFromIdent = "/* cypher */fromParDir"
    }

    if (toParentDirectoryId === "/") {
      matchToPath = "/* cypher */(root:UserRoot{ user: $client_username })"
      matchToIdent = "/* cypher */root"
    } else {
      matchToPath =
        "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(toParDir:Object{ id: $to_parent_dir_id })"
      matchToIdent = "/* cypher */toParDir"
    }

    let objectHasChildren: boolean

    {
      const res = await tx.run(
        `/* cypher */
					MATCH ${matchFromPath}

					RETURN EXISTS { (${matchFromIdent})-[:HAS_CHILD]->(:Object{ id: $object_id })-[:HAS_CHILD]->() } AS object_has_children
				`,
        {
          client_username: clientUsername,
          from_parent_dir_id: fromParentDirectoryId,
          object_id: objectId,
        }
      )

      if (!res.records.length) {
        return null
      }

      objectHasChildren = res.records[0]?.get("object_has_children")
    }

    let fileCopyIdMaps: { cloud_object_name: string; copy_id: string }[]

    if (objectHasChildren) {
      const res = await tx.run(
        `/* cypher */
				MATCH ${matchFromPath}
				MATCH (${matchFromIdent})-[:HAS_CHILD]->(obj:Object{ id: $object_id })
					((parents)-[:HAS_CHILD]->(children))+
	
				RETURN [p IN parents | p.id] AS parent_ids, [c IN children | c.id] AS children_ids
				`,
        {
          client_username: clientUsername,
          from_parent_dir_id: fromParentDirectoryId,
          object_id: objectId,
        }
      )

      if (!res.records.length) {
        return null
      }

      // last record contains the full parents and children
      const recObj = res.records.at(-1)?.toObject()

      const parentIds: string[] = recObj?.parent_ids
      const childrenIds: string[] = recObj?.children_ids

      const parentIdsLen = parentIds.length

      const parentIdToChildId: (string | undefined)[][] = []

      for (let i = 0; i < parentIdsLen; i++) {
        parentIdToChildId[i] = [parentIds[i], childrenIds[i]]
      }

      {
        await tx.run(
          `/* cypher */
					MATCH ${matchFromPath}

					UNWIND $par_id_to_child_id AS par_id_0_chi_id_1
					CALL (${matchFromIdent}, par_id_0_chi_id_1) {
						MATCH (${matchFromIdent})-[:HAS_CHILD]->+(par:Object{ id: par_id_0_chi_id_1[0] })

						MERGE (parentCopy:Object{ copied_id: par.id })
						ON CREATE
							SET parentCopy += par { .*, id: randomUUID(), native: null, date_created: $now, date_modified: $now }
			
						WITH par, par_id_0_chi_id_1, parentCopy

						MATCH (par)-[:HAS_CHILD]->(chi:Object{ id: par_id_0_chi_id_1[1] })

						CREATE (parentCopy)-[:HAS_CHILD]->(childCopy:Object{ copied_id: chi.id })
						SET childCopy += chi { .*, id: randomUUID(), date_created: $now, date_modified: $now }
					}`,
          {
            par_id_to_child_id: parentIdToChildId,
            client_username: clientUsername,
            from_parent_dir_id: fromParentDirectoryId,
            now,
          }
        )
      }

      {
        const res = await tx.run(
          `/* cypher */
					MATCH ${matchToPath}

					MATCH (obj:Object { copied_id: $object_id })

					CREATE (${matchToIdent})-[:HAS_CHILD]->(obj)

					WITH ${matchToIdent}, obj
					
					MATCH (obj)-[:HAS_CHILD]->*(cobj)

					WITH ${matchToIdent}, obj, cobj, 
						[o IN collect(obj) WHERE o.obj_type = "file" | o { .cloud_object_name, copy_id: o.id }] AS objFileCopyIdMaps,
						[co IN collect(cobj) WHERE co.obj_type = "file" | co { .cloud_object_name, copy_id: co.id }] AS cobjFileCopyIdMaps

					SET obj.copied_id = null,
						cobj.copied_id = null
					
					SET ${matchToIdent}.date_modified = $now

					RETURN objFileCopyIdMaps + cobjFileCopyIdMaps AS file_copy_id_maps
					`,
          {
            client_username: clientUsername,
            to_parent_dir_id: toParentDirectoryId,
            object_id: objectId,
            now: now,
          }
        )
        if (!res.records.length) {
          return null
        }

        fileCopyIdMaps = res.records[0]?.get("file_copy_id_maps")
      }
    } else {
      const res = await tx.run(
        `/* cypher */
				MATCH ${matchFromPath}
				MATCH ${matchToPath}

				MATCH (${matchFromIdent})-[:HAS_CHILD]->(obj:Object{ id: $object_id })

				CREATE (${matchToIdent})-[:HAS_CHILD]->(objCopy:Object)
				SET objCopy += obj { .*, id: randomUUID(), native: null, date_created: $now, date_modified: $now }

				SET ${matchToIdent}.date_modified = $now

				RETURN 
					CASE obj.obj_type 
						WHEN = "file" THEN [{ cloud_object_name: obj.cloud_object_name, copy_id: objCopy.id }]
						ELSE []
					END AS file_copy_id_maps
				`,
        {
          client_username: clientUsername,
          from_parent_dir_id: fromParentDirectoryId,
          to_parent_dir_id: toParentDirectoryId,
          object_id: objectId,
          now: now,
        }
      )
      if (!res.records.length) {
        return null
      }

      fileCopyIdMaps = res.records[0]?.get("file_copy_id_maps")
    }

    return fileCopyIdMaps
  })

  if (!res) {
    return { done: false, fileCopyIdMaps: [] }
  }

  return { done: true, fileCopyIdMaps: res }
}

export async function Mkfil(data: {
  clientUsername: string
  parentDirectoryId: string
  objectId: string
  cloudObjectName: string
  filename: string
  mimeType: string
  size: number
}) {
  let matchFromPath: string
  let matchFromIdent: string

  if (data.parentDirectoryId === "/") {
    matchFromPath = "/* cypher */(root:UserRoot{ user: $client_username })"
    matchFromIdent = "/* cypher */root"
  } else {
    matchFromPath =
      "/* cypher */(:UserRoot{ user: $client_username })-[:HAS_CHILD]->+(parObj:Object{ id: $parent_dir_id })"
    matchFromIdent = "/* cypher */parObj"
  }

  const res = await db.WriteQuery(
    `/* cypher */
			MATCH ${matchFromPath}
			CREATE (${matchFromIdent})-[:HAS_CHILD]->(obj:Object{ id: $object_id, obj_type: "file", name: $filename, cloud_object_name: $cloud_object_name, mime_type: $mime_type, size: $size, date_created: $now, date_modified: $now })
			
			RETURN obj { .* } AS new_file
		`,
    {
      client_username: data.clientUsername,
      parent_dir_id: data.parentDirectoryId,
      object_id: data.objectId,
      cloud_object_name: data.cloudObjectName,
      filename: data.filename,
      mime_type: data.mimeType,
      size: data.size,
      now: Date.now(),
    }
  )

  if (!res.records.length) {
    return null
  }

  return res.records[0]?.get("new_file") as FileT
}
