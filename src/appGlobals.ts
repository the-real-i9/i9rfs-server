import type { Storage } from "@google-cloud/storage"
import type { Driver } from "neo4j-driver"

let gcsClient: Storage

let neo4jDriver: Driver

export default {
  get Neo4jDriver() {
    return neo4jDriver
  },
  set Neo4jDriver(driver: Driver) {
    neo4jDriver = driver
  },
  get GCSClient() {
    return gcsClient
  },
  set GCSClient(client: Storage) {
    gcsClient = client
  },
}
