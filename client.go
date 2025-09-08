package main

import (
	"log"
	"time"

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
	// 메시지 수신 대기
	for {
		m, err := c.read()
		if err != nil {
			// 오류 생성시 수신 루프 종료
			log.Println("Read Msg Err: ", err)
			break
		}
		m.CreateMsg()
		broadcast(m)

	}

	c.Close()
}
func (c *Client) writeLoop() {
	// 클라이언트 수신대기

	for msg := range c.send {
		if c.roomId == msg.RoomID.Hex() {
			c.write(msg)
		}
	}
}

// 모든 클라이언트의 Send 채널에 전달
func broadcast(m *Message) {
	for _, client := range clients {
		client.send <- m
	}
}

func (c *Client) read() (*Message, error) {
	var msg *Message
	// WebSocket Conn에 json전달되면 Msg타입으로 메시지 읽음
	if err := c.conn.ReadJSON(&msg); err != nil {
		return nil, err
	}

	msg.CreatedAt = time.Now()

	msg.User = c.user

	log.Println("Read From WebSocket : ", msg)

	return msg, nil
}

func (c *Client) write(m *Message) error {
	log.Println("Write to WebSocket: ", m)

	return c.conn.WriteJSON(m)
}
