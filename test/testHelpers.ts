import { type TestContext } from "node:test"
import { type DirT } from "../src/appTypes.ts"

export function containsDirs(
  actual: DirT[],
  expectedDirs: string[],
  t: TestContext
) {
  for (const dirName of expectedDirs) {
    const dir = actual.find((d) => d.name === dirName)
    t.assert.ok(dir)
    t.assert.ok(dir.id)
    t.assert.strictEqual(dir.obj_type, "directory")
  }
}

export function notContainsDirs(
  actual: DirT[],
  notExpectedDirs: string[],
  t: TestContext
) {
  const actualDirs = actual.map((d) => d.name)

  for (const dirName of notExpectedDirs) {
    t.assert.ok(!actualDirs.includes(dirName))
  }
}
