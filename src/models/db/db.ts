import neo4j from "neo4j-driver"
import appGlobals from "../../appGlobals.ts"

export function WriteQuery(cypher: string, params: any) {
  return appGlobals.Neo4jDriver.executeQuery(cypher, params, {
    routing: neo4j.routing.WRITE,
  })
}

export function ReadQuery(cypher: string, params: any) {
  return appGlobals.Neo4jDriver.executeQuery(cypher, params, {
    routing: neo4j.routing.READ,
  })
}
