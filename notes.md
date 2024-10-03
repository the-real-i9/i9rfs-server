# Notes

## File System Management

> Mark when implemented

### Create directory

Command: `mkdir`

Flag: `-p`

Arg: `folderName`

#### Implementation

- Take the name of the folder, and with the name
- In the database, create a new "folder" and set the last segment of the work path as the parent folder. To set the last segment of the work path as the parent folder, it's either 
  - you have the parent folder's objectId or
  - you trace the folder from the root segment of the work path by name, since a parent-folder combination is unique (i.e. a folder can't have two children of the same name)

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

### Change Directory (Path exists test)

Command: `cd`
