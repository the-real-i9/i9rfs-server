package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

func ToStruct(val any, dest any) {
	if reflect.TypeOf(dest).Elem().Kind() != reflect.Struct {
		panic("expected 'dest' to be a pointer to struct")
	}

	bt, err := json.Marshal(val)
	if err != nil {
		log.Println("helpers.go: ToStruct: json.Marshal:", err)
	}

	if err := json.Unmarshal(bt, dest); err != nil {
		log.Println("helpers.go: ToStruct: json.Unmarshal:", err)
	}
}

func WSErrReply(err error, toAction string) map[string]any {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	errResp := map[string]any{
		"event":    "server error",
		"toAction": toAction,
		"data": map[string]any{
			"statusCode": errCode,
			"errorMsg":   fmt.Sprint(err),
		},
	}

	return errResp
}

func WSReply(data any, toAction string) map[string]any {

	reply := map[string]any{
		"event":    "server reply",
		"toAction": toAction,
		"data":     data,
	}

	return reply
}

func ValidationError(err error, filename, structname string) error {
	if err != nil {
		if e, ok := err.(validation.InternalError); ok {
			log.Printf("%s: %s: %v", filename, structname, e.InternalError())
			return fiber.ErrInternalServerError
		}

		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("validation error: %s", err))
	}

	return nil
}

func Cookie(name, value, path string, maxAge int) *fiber.Cookie {
	c := &fiber.Cookie{
		HTTPOnly: true,
		Secure:   false,
		Domain:   os.Getenv("SERVER_HOST"),
	}

	c.Name = name
	c.Value = value
	c.Path = path
	c.MaxAge = maxAge

	return c
}
