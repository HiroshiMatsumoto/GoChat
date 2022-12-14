package main

import (
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

func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteJSON(msg); err != nil {
			break
		}
	}
	c.socket.Close()
}

func (c *client) read() {
	// read method allows client to read from the socket.ReadJSON
	for {
		var msg *message
		if err := c.socket.ReadJSON(&msg); err == nil {
			msg.When = time.Now()
			msg.Name = c.userData["name"].(string)
			if avatarUrl, ok := c.userData["avatar_url"]; ok {
				msg.AvatarURL = avatarUrl.(string)
			}
			// send recieved msg to forward channel
			c.room.forward <- msg
		} else {
			// close socket
			break
		}
	}
	c.socket.Close()
}
