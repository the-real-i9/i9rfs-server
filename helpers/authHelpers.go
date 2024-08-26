package helpers

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func JwtSign(data any, secret string, expires time.Time) string {
	// create token -> (header.payload)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": data,
		"exp":  expires.Unix(),
	})

	// sign token with secret -> (header.payload.signature)
	jwt, err := token.SignedString([]byte(os.Getenv("SIGNUP_SESSION_JWT_SECRET")))
	if err != nil {
		panic(err)
	}

	return jwt
}

func JwtVerify[T any](tokenString, secret string) (*T, error) {
	parser := jwt.NewParser()
	token, err := parser.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		log.Println(err)
		return nil, err
	}

	var data *T

	MapToStruct(token.Claims.(jwt.MapClaims)["data"].(map[string]any), data)

	return data, nil
}
