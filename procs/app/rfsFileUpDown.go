package appprocs

import "os"

func (rfs RFS) UploadFile(args struct {
	WorkPath string
	Data     []byte
	Filename string
}, reply *string) error {

	if err := os.WriteFile(fsHome+args.WorkPath+"/"+args.Filename, args.Data, 0644); err != nil {
		return err
	}

	*reply = "Operation Successful"

	return nil
}

func (rfs RFS) DownloadFile(args struct {
	WorkPath string
	Filename string
}, reply *[]byte) error {

	data, err := os.ReadFile(fsHome + args.WorkPath + "/" + args.Filename)
	if err != nil {
		return err
	}

	*reply = data

	return nil
}
