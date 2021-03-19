package main

import (
	"fmt"
)

func cli_ui(srv *Server) {
	fmt.Printf("Scanner thread started\n")
	for {
		var in string
		fmt.Scan(&in)
		srv.broadcast <- in
	}
}

func main() {
	fmt.Println("Hello, world.")
	ws_srv := init_server(4)
	ws_srv.run(2)
	cli_ui(ws_srv)
	fmt.Println("A wiec to tak")
}
