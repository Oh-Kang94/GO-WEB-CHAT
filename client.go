package main

import (
	"log"

	"github.com/gorilla/websocket"
)

var clients []*Client

type Client struct {
	conn   *websocket.Conn // WebSocket Connection
	send   chan *Message   // 메시지 전송용 채널
	roomId string          // 현재 접속한 채팅방
	user   *User
}

const messageBufferSize = 256

func newCleint(conn *websocket.Conn, roomId string, u *User) {
	c := &Client{
		conn:   conn,
		send:   make(chan *Message, messageBufferSize),
		roomId: roomId,
		user:   u,
	}

	// 클라이언트 목록에 새로운 클라이언트 추가
	clients = append(clients, c)

	// 메시지 수신/전송 대기
	go c.readLoop()
	go c.writeLoop()
}

func (c *Client) Close() {
	for i, client := range clients {
		// 딱 해당하는 것만 빼기
		if client == c {
			clients = append(clients[:i], clients[i+1:]...)
		}
		break
	}
	// Send 채널 닫음
	close(c.send)

	// 웹소켓 종료
	c.conn.Close()

	log.Printf("close conn. addr: %s", c.conn.RemoteAddr())

}

func (c *Client) readLoop() {
	panic("unimplemented")
}

func (c *Client) writeLoop() {
	panic("unimplemented")
}
