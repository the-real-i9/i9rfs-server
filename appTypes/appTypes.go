package appTypes

import "time"

type SignupSessionData struct {
	Email                   string    `json:"email"`
	VerificationCode        int       `json:"verificationCode"`
	VerificationCodeExpires time.Time `json:"verificationCodeExpires"`
}

type SignupSession struct {
	Step string            `json:"step"`
	Data SignupSessionData `json:"data"`
}

type ClientUser struct {
	Username string `json:"username"`
}

type WSResp struct {
	StatusCode int    `json:"statusCode"`
	Body       any    `json:"body"`
	Error      string `json:"error"`
}
