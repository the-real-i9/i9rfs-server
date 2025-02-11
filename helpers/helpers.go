package helpers

import (
	"encoding/json"
	"i9rfs/appTypes"
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

func WSErrResp(err error) appTypes.WSResp {

	errCode := fiber.StatusInternalServerError

	if ferr, ok := err.(*fiber.Error); ok {
		errCode = ferr.Code
	}

	return appTypes.WSResp{
		StatusCode: errCode,
		ErrorMsg:   err.Error(),
	}
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
