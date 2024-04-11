package appcontrollers

import (
	"net/http"
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

func RFSCmd(w http.ResponseWriter, r *http.Request) {

}

type FSCmdArgs struct {
	WorkPath    string
	CmdLineArgs []string
}

func (rfs RFS) PathExists(args struct {
	WorkPath string
	Path     string
}, reply *string) error {
	_, err := os.ReadDir(fsHome + args.WorkPath + args.Path)
	if err != nil {
		*reply = "no"
		return nil
	}

	*reply = "yes"

	return nil
}

func (rfs RFS) CreateFile(args FSCmdArgs, reply *string) error {
	cmd := exec.Command("touch", args.CmdLineArgs...)
	cmd.Dir = fsHome + args.WorkPath

	err := cmd.Run()
	if err != nil {
		return err
	}

	*reply = ""
	return nil
}

func (rfs RFS) CreateDirectory(args FSCmdArgs, reply *string) error {
	cmd := exec.Command("mkdir", args.CmdLineArgs...)
	cmd.Dir = fsHome + args.WorkPath

	err := cmd.Run()
	if err != nil {
		return err
	}

	*reply = ""
	return nil
}

func (rfs RFS) RemoveDirectory(args FSCmdArgs, reply *string) error {
	cmd := exec.Command("rmdir", args.CmdLineArgs...)
	cmd.Dir = fsHome + args.WorkPath

	err := cmd.Run()
	if err != nil {
		return err
	}

	*reply = ""
	return nil
}

func (rfs RFS) Remove(args FSCmdArgs, reply *string) error {
	cmd := exec.Command("rm", args.CmdLineArgs...)
	cmd.Dir = fsHome + args.WorkPath

	err := cmd.Run()
	if err != nil {
		return err
	}

	*reply = ""
	return nil
}

func (rfs RFS) Copy(args FSCmdArgs, reply *string) error {
	cmd := exec.Command("cp", args.CmdLineArgs...)
	cmd.Dir = fsHome + args.WorkPath

	err := cmd.Run()
	if err != nil {
		return err
	}

	*reply = ""
	return nil
}

func (rfs RFS) MoveRename(args FSCmdArgs, reply *string) error {
	cmd := exec.Command("mv", args.CmdLineArgs...)
	cmd.Dir = fsHome + args.WorkPath

	err := cmd.Run()
	if err != nil {
		return err
	}

	*reply = ""
	return nil
}

func (rfs RFS) PrintContent(args FSCmdArgs, reply *string) error {
	strb := &strings.Builder{}

	cmd := exec.Command("cat", args.CmdLineArgs...)
	cmd.Dir = fsHome + args.WorkPath
	cmd.Stdout = strb

	err := cmd.Run()
	if err != nil {
		return err
	}

	*reply = strb.String()
	return nil
}

func (rfs RFS) ListDirectoryContents(args FSCmdArgs, reply *string) error {
	strb := &strings.Builder{}

	cmd := exec.Command("ls", args.CmdLineArgs...)
	cmd.Dir = fsHome + args.WorkPath
	cmd.Stdout = strb

	err := cmd.Run()
	if err != nil {
		return err
	}

	*reply = strb.String()
	return nil
}
