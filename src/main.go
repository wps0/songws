package main

import (
	"flag"
	"fmt"
	"runtime"
)

func cli_ui(srv *Server) {
	fmt.Printf("Scanner thread started\n")
	for {
		var in string
		fmt.Scan(&in)
		fmt.Printf("GoRoutines number: %d; clients len: %d\n", runtime.NumGoroutine(), len(srv.clients))
		srv.broadcast <- in
	}
}

var ip = flag.String("ip", "0.0.0.0", "The IPv4 address on which the server should listen")
var port = flag.Int("p", 8080, "The port of the server")
var cc = flag.Int("c", 1024, "Maximum number of clients connected concurrently to the server.\nIf a client tries connecting after the limit was reached, a 503 error will be returned.")

func main() {
	flag.Parse()

	ws_srv := init_server(*cc)
	ws_srv.run(2)
	go cli_ui(ws_srv)
	init_http(ws_srv, *ip, *port)
	fmt.Println("A wiec to tak")
}
