package rfsCmdService

import (
	"fmt"
	"i9rfs/server/models/rfsCmdModel"
	"strings"
)

func resolveToTargetPath(currentWorkPath, targetPath string) string {
	dirs := strings.Split(targetPath, "/")

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

func ChangeDirectory(workPath string, cmdArgs []string) (string, error) {
	targetPath := cmdArgs[0]
	if strings.HasPrefix(targetPath, "/") {
		return "", fmt.Errorf("invalid target path %s. Did you mean to prefix with ./ instead?", targetPath)
	}

	resolvedPath := resolveToTargetPath(workPath, targetPath)

	if resolvedPath == "/" {
		return resolvedPath, nil
	}

	exists, err := rfsCmdModel.PathExists(resolvedPath)
	if err != nil {
		return "", err
	}

	if !exists {
		return "", fmt.Errorf("no such file or directory")
	}

	return resolvedPath, nil
}

func MakeDirectory(workPath string, cmdArgs []string) (string, error) {

	return "Operation Successful", nil
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
