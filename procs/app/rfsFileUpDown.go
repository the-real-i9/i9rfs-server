package appprocs

type FSUpDownArgs struct {
	WorkPath string
	Data     []byte
	Filename string
}

func (rfs *RFS) UploadFile(args *FSUpDownArgs, reply *string) error {

	return nil
}

func (rfs *RFS) DownloadFile(args *FSUpDownArgs, reply *[]byte) error {

	return nil
}
