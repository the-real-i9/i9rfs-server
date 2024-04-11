package rfscmdservice

import (
	"os"
	"os/exec"
	"strings"
)

var fsHome = "i9FSHome"

func init() {
	if hdir, err := os.UserHomeDir(); err == nil {
		fsHome = hdir + "/i9FSHome"
	}
}

func PathExists(workPath string) (bool, error) {
	f, err := os.Open(fsHome + workPath)
	if err != nil {
		return false, nil
	}

	defer f.Close()
	return true, nil
}

func FileMgmtCommand(workPath string, command string, cmdArgs []string) (string, error) {
	strb := &strings.Builder{}

	cmd := exec.Command(command)
	cmd.Dir = fsHome + workPath
	cmd.Stdout = strb

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strb.String(), nil
}

func UploadFile(workPath string, cmdArgs []string) (string, error) {
	data := []byte(cmdArgs[0])
	filename := cmdArgs[1]

	if err := os.WriteFile(fsHome+workPath+"/"+filename, data, 0644); err != nil {
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
