package signupControllers

import (
	"context"
	"i9rfs/server/appTypes"
	"i9rfs/server/helpers"
	"i9rfs/server/services/signupService"

	"github.com/gofiber/fiber"
)

func verifyEmail(ctx context.Context, sessionId string, sessionData *appTypes.SignupSessionData, data map[string]any) appTypes.WSResp {
	var body verifyEmailBody

	helpers.MapToStruct(data, &body)

	if val_err := body.Validate(); val_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, val_err)
	}

	respData, app_err := signupService.VerifyEmail(ctx, sessionId, sessionData.VerfCode, body.Code, sessionData.Email)

	if app_err != nil {
		return helpers.ErrResp(fiber.StatusUnprocessableEntity, app_err)
	}

	return appTypes.WSResp{
		StatusCode: fiber.StatusOK,
		Body:       respData,
	}
}
