package signupControllers

import (
	"context"
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"i9rfs/server/services/signupService"

	"github.com/gofiber/fiber"
)

func registerUser(ctx context.Context, sessionId string, sessionData *appTypes.SignupSessionData, data map[string]any) appTypes.WSResp {
	var body registerUserBody

	helpers.MapToStruct(data, &body)

	if val_err := body.Validate(); val_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err)
	}

	respData, app_err := signupService.RegisterUser(ctx, sessionId, sessionData.Email, body.Username, body.Password)

	if app_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err)
	}

	return appTypes.WSResp{
		StatusCode: fiber.StatusOK,
		Body:       respData,
	}
}
