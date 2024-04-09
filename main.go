package main

import (
	"fmt"
	"os"
)

/* "i9Packages/i9helpers"
"i9rfs/procs/appprocs"
"i9rfs/procs/authprocs"
"log"
"net"
"net/http"
"net/rpc" */

func main() {
	/* if err := i9helpers.LoadEnv(".env"); err != nil {
		log.Fatal(err)
	}

	if err := i9helpers.InitDBPool(); err != nil {
		log.Fatal(err)
	}

	authSignup := new(authprocs.AuthSignup)
	auth := new(authprocs.Auth)
	rfs := new(appprocs.RFS)

	rpc.Register(auth)
	rpc.Register(authSignup)
	rpc.Register(rfs)

	rpc.HandleHTTP()

	listn, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}

	go http.Serve(listn, nil) */

	fmt.Println(os.UserHomeDir())
}
