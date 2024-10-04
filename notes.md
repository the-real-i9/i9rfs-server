# Notes

## File System Management

> Mark when implemented

### Create directory

Command: `mkdir`

Flag: `-p`

Arg: `folderName`

#### Implementation

- Take the `dirName` of the directory
- Create a new directory with the `dirName` in the database
- If the `workPath` is "/" then the parent_directory_id is empty, else parent directory is the id of the directory with the `path` of `workPath`.
- The path attribute of the new directory is `workPath` + `/dirName` (or `workPath` + `dirName`, if `workPath` is `/`)

### Remove directory

Command: `rmdir`

Flag: `-r`

Arg: `folderName`

### Copy

Command: `cp`

Arg: `fileOrFolderName`

### Move/Rename

Command: `mv`

Arg: `fileOrFolderName`

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

