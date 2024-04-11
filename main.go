package main

import (
	"i9pkgs/i9helpers"
	"i9rfs/server/routes/approutes"
	"i9rfs/server/routes/authroutes"

	"log"
	"net/http"
)

func main() {
	if err := i9helpers.LoadEnv(".env"); err != nil {
		log.Fatal(err)
	}

	if err := i9helpers.InitDBPool(); err != nil {
		log.Fatal(err)
	}

	authroutes.Init()
	approutes.Init()

	go http.ListenAndServe(":8000", nil)
}
