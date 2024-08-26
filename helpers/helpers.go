package helpers

import (
	"encoding/json"
	"errors"
	"i9rfs/server/appTypes"
	"i9rfs/server/globalVars"
)

func MapToStruct(val map[string]any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ToStruct(val any, structData any) {
	bt, _ := json.Marshal(val)

	json.Unmarshal(bt, structData)
}

func ErrResp(code int, err error) appTypes.WSResp {
	if errors.Is(err, globalVars.ErrInternalServerError) {
		return appTypes.WSResp{StatusCode: 500, Error: globalVars.ErrInternalServerError.Error()}
	}

	return appTypes.WSResp{StatusCode: code, Error: err.Error()}
}
