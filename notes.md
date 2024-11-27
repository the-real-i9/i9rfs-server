# Notes

## File System Management

### Copy

Command: `cp`

Arg: `fileOrFolderName`

### Move/Rename

Command: `mv`

Arg: `SOURCE DEST`

#### Implememtation

- If: `SOURCE` path exists,
  - Get the `id` of `SOURCE`
- Else: throw an error
- If: `DEST` path, "excluding the last segment", exists
  - If the last segment exists as a directory
    - Set the `id` of `DEST` as the `parent_directory_id` of `SOURCE`
  - Else: Set the `id` of `DEST` "excluding the last segment" as the `parent_directory_id` of `SOURCE`, then rename the first segment of `SOURCE` path to the name of the last segment of `DEST` path.
- Else: throw an error

### List directory contents

Command: `ls`

Flags: `-s PROP` - sort by property, `-d` - show details (properties)

Arg: `[fileOrFolderName]`

### File/Directory Properties

Command: `props`

Arg: `pathToFileOrFolder`

### Download

Command: `download`

Arg: `pathToFile`

### Upload

Command: `upload`

Arg: `pathToFile`
