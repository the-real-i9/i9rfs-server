package helpers

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func JwtSign(data any, secret string, expires time.Time) string {
	// create token -> (header.payload)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data":  data,
		"admin": true,
		"exp":   expires,
	})

	// sign token with secret -> (header.payload.signature)
	jwt, err := token.SignedString([]byte(os.Getenv("SIGNUP_SESSION_JWT_SECRET")))
	if err != nil {
		panic(err)
	}

	return jwt
}
