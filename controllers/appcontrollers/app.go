package appcontrollers

import "nhooyr.io/websocket"

var wsOpts = &websocket.AcceptOptions{OriginPatterns: []string{"localhost"}}
