package appprocs

import (
	"fmt"
	"os/exec"
)

type FSCmdArgs struct {
	WorkPath    string
	CmdLineArgs []string
}

func (rfs *RFS) CreateDirectory(args *FSCmdArgs, reply *string) error {
	cmd := exec.Command("mkdir", args.CmdLineArgs...)
	cmd.Dir = fmt.Sprintf("i9FSHome%s", args.WorkPath)

	err := cmd.Run()
	if err != nil {
		return err
	}

	*reply = ""
	return nil
}
