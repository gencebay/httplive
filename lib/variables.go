package lib

import "github.com/gorilla/websocket"

// Environments ...
var Environments = EnvironmentVariables{DatabaseName: "httplive.db"}

// DefaultMemory Form data ...
const DefaultMemory = 32 * 1024 * 1024

// Clients ...
var Clients = make(map[string]*websocket.Conn)
