package rfsCmdService

import (
	"context"
	"fmt"
	"i9rfs/models/rfsCmdModel"
	"strings"
)

func resolveToTarget(currentWorkPath, target string) string {
	if strings.HasPrefix(target, "/") {
		return target
	}

	dirs := strings.Split(target, "/")

	newWorkPath := currentWorkPath

	for _, dir := range dirs {
		if dir == "." {
			continue
		} else if dir == ".." {
			if newWorkPath == "/" {
				// the user has specified an invalid directory,
				// one that possibly tries to go out of their user account directory
				continue
			}

			// strip the last dir
			// if newWorkPath was the last directory in the root
			// the code line below will make it an empty string
			newWorkPath = newWorkPath[0:strings.LastIndex(newWorkPath, "/")]
			// hence, we check and restore to root
			if newWorkPath == "" {
				newWorkPath = "/"
			}
		} else {
			// append the dir
			if newWorkPath == "/" { // root
				newWorkPath += dir
			} else { // non-root
				newWorkPath += "/" + dir
			}
		}
	}

	return newWorkPath
}

func ChangeDirectory(ctx context.Context, workPath string, cmdArgs []string) (string, error) {
	resolvedPath := resolveToTarget(workPath, cmdArgs[0])

	if resolvedPath == "/" {
		return resolvedPath, nil
	}

	exists, err := rfsCmdModel.PathExists(ctx, resolvedPath)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("no such file or directory")
	}

	return resolvedPath, nil
}

func MakeDirectory(ctx context.Context, workPath string, cmdArgs []string, username string) (bool, error) {
	newDir := cmdArgs[0]

	if newDir == "/" {
		return false, fmt.Errorf("cannot create directory '/': File exists")
	}

	return rfsCmdModel.Mkdir(ctx, workPath, strings.Split(cmdArgs[0], "/"), username)
}

func RemoveDirectory(ctx context.Context, workPath string, cmdArgs []string, clientUsername string) (bool, error) {
	resolvedDirPath := resolveToTarget(workPath, cmdArgs[0])

	return rfsCmdModel.Rmdir(ctx, resolvedDirPath, cmdArgs[0], clientUsername)
}

func deleteFilesInCS(fileIds []string) {

}

func Remove(ctx context.Context, workPath string, cmdArgs []string, clientUsername string) (bool, error) {
	var (
		recursive        bool
		objectPathCmdArg string
	)

	if cmdArgs[0] != "-r" {
		recursive = false
		objectPathCmdArg = cmdArgs[0]
	} else {
		recursive = true
		objectPathCmdArg = cmdArgs[1]
	}

	objectPath := resolveToTarget(workPath, cmdArgs[0])
	status, fileIds, err := rfsCmdModel.Rm(ctx, objectPath, recursive, objectPathCmdArg, clientUsername)

	go deleteFilesInCS(fileIds)

	return status, err
}

func Move(workPath string, cmdArgs []string) (bool, error) {
	sourcePath := resolveToTarget(workPath, cmdArgs[0])
	destPath := resolveToTarget(workPath, cmdArgs[1])

	if sourcePath == "/" {
		return false, fmt.Errorf("cannot move '$source' to '$dest/$source': Device or resource busy")
	}

	return rfsCmdModel.Mv(sourcePath, destPath)

	// the .Mv model must tell you if you need to do a renaming on the GCS cloud
	// if sourcePath and destPath are both files, it must return the id of both
	// so you can find the source (by its id) on GCS and rename it to dest's id
}

func UploadFile(workPath string, cmdArgs []string) (string, error) {
	// fileData := []byte(cmdArgs[0])
	// filename := cmdArgs[1]

	// upload file to GCS

	return "Operation Successful", nil
}

func DownloadFile(workPath string, cmdArgs []string) ([]byte, error) {
	// filename := cmdArgs[0]

	// retrieve file from GCS

	return nil, nil
}
