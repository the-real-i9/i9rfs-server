package authroutes

import (
	authcontrollers "i9rfs/server/controllers/auth"
	"net/http"
)

func Init() {
	http.HandleFunc("/api/auth/signup", authcontrollers.Signup)

	http.HandleFunc("/api/auth/login", authcontrollers.Login)
}
