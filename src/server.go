package main

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

const CLIENT_BUFFER_SIZE = 4

type StatusTrack struct {
	Artist    string `json:"artist"`
	Song      string `json:"title"`
	Streaming bool   `json:"streaming"`
	Date      int    `json:"date"`
}

func track_to_status_track(t Track) StatusTrack {
	now_streaming, _ := strconv.Atoi(t.Streamable)
	date, _ := strconv.Atoi(t.Date.Uts)
	return StatusTrack{
		Artist:    t.Artist.Text,
		Song:      t.Name,
		Streaming: now_streaming > 0,
		Date:      date,
	}

}

type StatusUpdate struct {
	// 0 - status update
	// 1 - new song
	Status int           `json:"msg_type"`
	Data   []StatusTrack `json:"data"`
}

type Server struct {
	clients     map[*Client]bool
	max_clients int
	// messages to be broadcasted to all clients
	broadcast chan string
	// client management
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func init_server(max_clients int) *Server {
	return &Server{
		max_clients: max_clients,
		clients:     make(map[*Client]bool),
		broadcast:   make(chan string),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
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
			srv.mu.Lock()
			srv.clients[cli] = true
			srv.mu.Unlock()
		case cli := <-srv.unregister:
			srv.mu.Lock()
			delete(srv.clients, cli)
			close(cli.comm)
			srv.mu.Unlock()
		case bc := <-srv.broadcast:
			srv.mu.RLock()
			for k := range srv.clients {
				k.comm <- []byte(bc)
			}
			srv.mu.RUnlock()
		}
	}
}

func (srv *Server) does_accept_clients() bool {
	defer srv.mu.RUnlock()
	srv.mu.RLock()
	return len(srv.clients) < srv.max_clients
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:   CLIENT_BUFFER_SIZE,
	WriteBufferSize:  CLIENT_BUFFER_SIZE,
	HandshakeTimeout: 3000,
}

func ws_handler(srv *Server, rw http.ResponseWriter, r *http.Request) {
	if !srv.does_accept_clients() {
		log.Printf("[Websocket server] Connection from %s refused: maximum number of clients (%d) reached\n", r.RemoteAddr, srv.max_clients)
		rw.WriteHeader(503)
		return
	}
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		addr := "unknown"
		if conn != nil {
			addr = conn.RemoteAddr().String()
		}
		log.Printf("[Websocket server] Connection attempt from %s failed with error: %s", addr, err)
		return
	}

	log.Printf("[Websocket server] Connection from %s to %s (User-Agent: %s)\n", r.RemoteAddr, r.URL, r.UserAgent())
	client := create_client(conn, srv)
	go client_writer(client)
	go client_reader(client)
}

func init_http(srv *Server, ip string, port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/home/piotr/Documents/golang/songws/src/home.html")
	})
	http.HandleFunc("/ws", func(rw http.ResponseWriter, r *http.Request) {
		ws_handler(srv, rw, r)
	})

	if len(*cert) > 0 && len(*pem) > 0 && *en_https {
		log.Println("Enabling HTTPS server...")
		go log.Fatal(http.ListenAndServeTLS(ip+":"+strconv.Itoa(port), *cert, *pem, nil))
	}

	if *en_http {
		log.Println("Enabling HTTP server...")
		go log.Fatal(http.ListenAndServe(ip+":"+strconv.Itoa(port), nil))
	}
}
