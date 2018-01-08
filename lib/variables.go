package lib

import "github.com/gorilla/websocket"

// Environments ...
var Environments = EnvironmentVariables{DatabaseName: "httplive.db"}

// Clients ...
var Clients = make(map[*websocket.Conn]bool)

// Broadcast ...
var Broadcast = make(chan WsMessage)
