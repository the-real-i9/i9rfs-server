package rfsCmdService

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

var fsHome = ""

func SetHome(homePath string) {
	fsHome = homePath
}

func PathExists(workPath string) (bool, error) {
	f, err := os.Open(fsHome + workPath)
	if err != nil {
		return false, nil
	}

	defer f.Close()
	return true, nil
}

func BashCommand(workPath string, command string, cmdArgs []string) (string, error) {
	res := &strings.Builder{}
	errRes := &strings.Builder{}

	cmd := exec.Command(command, cmdArgs...)
	cmd.Dir = fsHome + workPath
	cmd.Stdout = res
	cmd.Stderr = errRes

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s", errRes.String())
	}

	return res.String(), nil
}

func UploadFile(workPath string, cmdArgs []string) (string, error) {
	fileData := []byte(cmdArgs[0])
	filename := cmdArgs[1]

	if err := os.WriteFile(fsHome+workPath+"/"+filename, fileData, fs.ModePerm); err != nil {
		return "", err
	}

	return "Operation Successful", nil
}

func DownloadFile(workPath string, cmdArgs []string) ([]byte, error) {
	filename := cmdArgs[0]

	data, err := os.ReadFile(fsHome + workPath + "/" + filename)
	if err != nil {
		return nil, err
	}

	return data, nil
}
