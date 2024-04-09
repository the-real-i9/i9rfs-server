package appprocs

type FSUpDownParams struct {
	WorkPath string
	Data     []byte
	Filename string
}

func (rfs *RFS) UploadFile(args *FSUpDownParams, reply *string) error {

	return nil
}

func (rfs *RFS) DownloadFile(args *FSUpDownParams, reply *[]byte) error {

	return nil
}
