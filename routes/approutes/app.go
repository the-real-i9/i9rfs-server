package approutes

import (
	"i9rfs/server/controllers/appcontrollers"
	"net/http"
)

func Init() {
	http.HandleFunc("/api/app/get_session_user", appcontrollers.GetSessionUser)
	http.HandleFunc("/api/app/rfs", appcontrollers.RFSCmd)
}
