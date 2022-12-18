package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	// socket is the web socket for this client
	socket *websocket.Conn
	// send is the web socket for this client
	send chan *message
	// room is the room this client is chatting in
	room *room
	// user information
	userData map[string]interface{}
}

func (c *client) read() {
	log.Println("client.read()")
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			// log.Printf("message: %s by %s", msg.Message, msg.Name)
			c.room.forward <- msg // send msg to room.forward
		} else {
			break
		}
	}
	c.socket.Close()
}

func (c *client) write() {
	// log.Println("client.write()")
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
