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
	// room to join
	join chan *client
	// leave room
	leave chan *client
	// all clients in the room
	clients map[*client]bool
	//
	tracer trace.Tracer
}

// newRoom makes a new room that is ready to go
func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		tracer:  trace.Off(),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
			r.tracer.Trace("New client joined")
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("Client left")
			log.Print("Client left")
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
	//
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	authCookie, err := req.Cookie("auth")
	if err != nil {
		log.Fatal("Failed to get auth cookie:", err)
		return
	}

	client := &client{
		socket:   socket,
		send:     make(chan *message, messageBufferSize), // message objects are send/recieved
		room:     r,
		userData: objx.MustFromBase64(authCookie.Value), // decode
	}

	// ??
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}
