package main

import (
	"github.com/gorilla/websocket"
)

type client struct {
	// socket is the web socket for this client
	socket *websocket.Conn
	// send is the web socket for this client
	send chan []byte
	// room is the room this client is chatting in
	room *room
}

func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.forward <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
