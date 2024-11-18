package securityServices

import (
	"errors"
	"fmt"
	"i9rfs/server/helpers"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func HashAndPasswordMatches(hash, plainPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

func GetTokenAndExpiration() (int, time.Time) {
	return rand.Intn(899999) + 100000, time.Now().UTC().Add(1 * time.Hour)
}

func JwtSign(data any, secret string, expires time.Time) string {
	// create token -> (header.payload)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": data,
		"exp":  expires.Unix(),
	})

	// sign token with secret -> (header.payload.signature)
	jwt, err := token.SignedString([]byte(secret))
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
		return nil, err
	}

	var data T

	helpers.MapToStruct(token.Claims.(jwt.MapClaims)["data"].(map[string]any), &data)

	return &data, nil
}
