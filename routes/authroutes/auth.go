package authroutes

import (
	"i9rfs/server/controllers/authcontrollers"
	"net/http"
)

func Init() {
	http.HandleFunc("/api/auth/signup", authcontrollers.Signup)

	http.HandleFunc("/api/auth/login", authcontrollers.Login)
}
