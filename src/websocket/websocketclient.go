package websocket

import "github.com/gorilla/websocket"

// a new web socket client consists of a web socket connection, and a user
// a user has ONLY ONE connection
type WSClient struct {
	Connection *websocket.Conn
	Interface  interface{}
}