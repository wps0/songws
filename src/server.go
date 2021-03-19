package main

import "fmt"

type Client struct {
}

type Server struct {
	clients []*Client
	// broadcast requests to the server
	broadcast chan string
	// client management
	register   chan *Client
	unregister chan *Client
}

func init_server(max_clients int) *Server {
	return &Server{
		clients:    make([]*Client, 0, max_clients),
		broadcast:  make(chan string),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (srv *Server) run(threads int) {
	for i := 0; i < threads; i++ {
		go srv.handler()
	}
}

func (srv *Server) handler() {
	for {
		select {
		case cli := <-srv.register:
			srv.clients = append(srv.clients, cli)
		case cli := <-srv.unregister:
			fmt.Printf("OK %v\n", cli)
		case bc := <-srv.broadcast:
			fmt.Printf("Broadcast message: %s\n", bc)
		}
	}
}
