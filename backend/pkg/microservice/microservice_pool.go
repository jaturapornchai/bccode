package microservice

import (
	"sync"

	"github.com/gorilla/websocket"
)

type CachePool struct {
	sync.Mutex
}

type WebsocketPool struct {
	Handler websocket.Upgrader
	sync.Mutex
	Connections map[string]*websocket.Conn
}
