package main

import (
	"github.com/gorilla/websocket"
)

var clients []*Client

type Client struct {
	conn   *websocket.Conn // WebSocket Connection
	send   chan *Message   // 메시지 전송용 채널
	roomId string          // 현재 접속한 채팅방
	user   *User
}
