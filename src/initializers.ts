import neo4j from "neo4j-driver"
import { Storage } from "@google-cloud/storage"
import appGlobals from "./appGlobals.ts"

function initGCSClient() {
  appGlobals.GCSClient = new Storage({
    apiKey: process.env.GCS_API_KEY || "",
  })
}

async function initNeo4jDriver() {
  const driver = neo4j.driver(
    process.env.NEO4J_URL || "",
    neo4j.auth.basic(
      process.env.NEO4J_USER || "",
      process.env.NEO4J_PASSWORD || ""
    )
  )

  const sess = driver.session()

  await sess.executeWrite(async (tx) => {
    await tx.run(
      `/* cypher */CREATE CONSTRAINT unique_username IF NOT EXISTS FOR (u:User) REQUIRE u.username IS UNIQUE`,
      null
    )

    await tx.run(
      `/* cypher */CREATE CONSTRAINT unique_email IF NOT EXISTS FOR (u:User) REQUIRE u.email IS UNIQUE`,
      null
    )

    await tx.run(
      `/* cypher */CREATE CONSTRAINT unique_object IF NOT EXISTS FOR (o:Object) REQUIRE o.id IS UNIQUE`,
      null
    )

    await tx.run(
      `/* cypher */CREATE CONSTRAINT unique_object_copy IF NOT EXISTS FOR (oc:Object) REQUIRE oc.copied_id IS UNIQUE`,
      null
    )
  })

  await sess.close()

  appGlobals.Neo4jDriver = driver
}

export async function InitApp() {
  initGCSClient()

  await initNeo4jDriver()
}

export function CleanUp() {
  appGlobals.Neo4jDriver.close()
}
