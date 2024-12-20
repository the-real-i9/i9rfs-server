package appTypes

type SignupSessionData struct {
	Step     string
	Email    string
	VerfCode int
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
