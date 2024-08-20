package appTypes

type SignupSessionData struct {
	SessionId string
	Email     string
	State     string
}

type ClientUser struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}

type WSResp struct {
	StatusCode int    `json:"statusCode"`
	Body       any    `json:"body"`
	Error      string `json:"error"`
}
