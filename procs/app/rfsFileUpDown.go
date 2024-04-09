package appprocs

func (rfs RFS) UploadFile(args struct {
	WorkPath string
	Data     []byte
	Filename string
}, reply *string) error {

	return nil
}

func (rfs RFS) DownloadFile(args struct {
	WorkPath string
	Filename string
}, reply *[]byte) error {

	return nil
}
