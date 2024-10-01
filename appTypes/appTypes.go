package appTypes

type SignupSessionData struct {
	SessionId string
	Email     string
	Step      string
}

type ClientUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
}

type WSResp struct {
	StatusCode int    `json:"statusCode"`
	Body       any    `json:"body"`
	Error      string `json:"error"`
}
