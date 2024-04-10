package appprocs

import "os"

type RFS struct{}

var fsHome = "i9FSHome"

func init() {
	if hdir, err := os.UserHomeDir(); err == nil {
		fsHome = hdir + "/i9FSHome"
	}
}
