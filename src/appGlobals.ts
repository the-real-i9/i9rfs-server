import type { Bucket } from "@google-cloud/storage"
import type { Driver } from "neo4j-driver"

let appGCSBucket: Bucket

let neo4jDriver: Driver

export default {
  get Neo4jDriver() {
    return neo4jDriver
  },
  set Neo4jDriver(driver: Driver) {
    neo4jDriver = driver
  },
  get AppGCSBucket() {
    return appGCSBucket
  },
  set AppGCSBucket(bucket: Bucket) {
    appGCSBucket = bucket
  },
}
