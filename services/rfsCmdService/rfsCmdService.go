package rfsCmdService

func ChangeDirectory(workPath string, cmdArgs []string) (string, error) {

	return "", nil
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
