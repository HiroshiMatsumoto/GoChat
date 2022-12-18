package main

import (
	"chat/trace"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/stretchr/objx"
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients
	forward chan *message
	//
	join chan *client
	//
	leave chan *client
	//
	clients map[*client]bool
	//
	tracer trace.Tracer
}

// newRoom makes a new room that is ready to go
func newRoom() *room {
	return &room{
		// create forward channel
		forward: make(chan *message),
		// create join channel
		join: make(chan *client),
		// create leave channel
		leave: make(chan *client),
		// create clients channel
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	// infinite loop
	for {
		// select statement: select randomly if multiple matching cases exist
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
		case msg := <-r.forward:
			r.tracer.Trace("Message received: ", msg.Message)
			for client := range r.clients {
				select {
				case client.send <- msg:
					// send the message
					r.tracer.Trace(" -- sent to client")
				default:
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace(" -- failed to send, cleaned up client")
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// a room acts as a handler

	// upgrader.Upgrade method recieves the socket
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	// auth check
	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}

	// create client
	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize), // message objects are send/recieved
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value), // decode
	}

	// pass it into the join channel
	r.join <- client
	defer func() { r.leave <- client }()
	// run the method in different thereads
	go client.write()
	// block operations and close
	client.read()
}
