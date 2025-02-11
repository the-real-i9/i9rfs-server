package helpers

import (
	"encoding/json"
	"errors"
	"i9rfs/server/appGlobals"
	"i9rfs/server/appTypes"
	"log"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gofiber/fiber/v2"
)

func MapToStruct(val map[string]any, yourStruct any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, yourStruct)
}

func StructToMap(val any, yourMap *map[string]any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, yourMap)
}

func ToStruct(val any, yourStruct any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, yourStruct)
}

func ErrResp(code int, err error) appTypes.WSResp {
	if errors.Is(err, appGlobals.ErrInternalServerError) {
		return appTypes.WSResp{StatusCode: 500, Error: appGlobals.ErrInternalServerError.Error()}
	}

	return appTypes.WSResp{StatusCode: code, Error: err.Error()}
}

func ValidationError(err error, filename, structname string) error {
	if err != nil {
		if e, ok := err.(validation.InternalError); ok {
			log.Printf("%s: %s: %v", filename, structname, e.InternalError())
			return fiber.ErrInternalServerError
		}

		return fiber.NewError(fiber.StatusBadRequest, "validation error:", err.Error())
	}

	return nil
}
