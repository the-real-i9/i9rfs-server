package helpers

import (
	"encoding/json"
	"errors"
	"i9rfs/server/appGlobals"
	"i9rfs/server/appTypes"
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
