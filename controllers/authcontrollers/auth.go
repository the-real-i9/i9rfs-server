package authcontrollers

import "nhooyr.io/websocket"

var wsOpts = &websocket.AcceptOptions{OriginPatterns: []string{"localhost"}}
