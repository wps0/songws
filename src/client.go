package main

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	server *Server
	comm   chan []byte
	ws_con *websocket.Conn
}

func create_client(conn *websocket.Conn, srv *Server) *Client {
	cli := &Client{
		comm:   make(chan []byte, CLIENT_BUFFER_SIZE),
		ws_con: conn,
		server: srv,
	}
	srv.register <- cli
	return cli
}

func client_writer(client *Client) {
	err := client.ws_con.WriteMessage(websocket.TextMessage, []byte(dq.To_Json()))
	if err != nil {
		Log.Printf("[Websocket client %s] Write error: %s", client.ws_con.RemoteAddr(), err)
		return
	}

	for {
		select {
		case msg := <-client.comm:
			err = client.ws_con.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				Log.Printf("[Websocket client %s] Write error: %s", client.ws_con.RemoteAddr(), err)
				return
			}
		}
	}
}

func client_reader(client *Client) {
	defer func() {
		client.server.unregister <- client
		client.ws_con.Close()
	}()

	for {
		_, _, err := client.ws_con.ReadMessage()
		if err != nil {
			Log.Printf("[Websocket client %s] Read error: %s", client.ws_con.RemoteAddr(), err)
			return
		}
		Log.Printf("[Websocket client %s] Illegal write attempt. Connection will be closed", client.ws_con.RemoteAddr())
	}
}
