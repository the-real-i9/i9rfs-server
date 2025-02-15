package securityServices

import (
	"errors"
	"fmt"
	"i9rfs/helpers"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
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

func PasswordMatchesHash(hash, plainPassword string) (bool, error) {
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
	var token int
	expires := time.Now().UTC().Add(1 * time.Hour)

	if os.Getenv("GO_ENV") != "production" {
		token, _ = strconv.Atoi(os.Getenv("DUMMY_VERF_TOKEN"))
	} else {
		token = rand.Intn(899999) + 100000
	}

	return token, expires
}

func JwtSign(data any, secret string, expires time.Time) (string, error) {
	// create token -> (header.payload)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": data,
		"exp":  expires.Unix(),
	})

	// sign token with secret -> (header.payload.signature)
	jwt, err := token.SignedString([]byte(secret))

	if err != nil {
		log.Println("securityServices.go: JwtSign:", err)
		return "", fiber.ErrInternalServerError
	}

	return jwt, err
}

func JwtVerify[T any](tokenString, secret string) (T, error) {
	var data T

	parser := jwt.NewParser()
	token, err := parser.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {

		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		return data, err
	}

	helpers.ToStruct(token.Claims.(jwt.MapClaims)["data"], &data)

	return data, nil
}
