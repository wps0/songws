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
var cc = flag.Int("c", 1024, "Maximum number of clients connected concurrently to the server.")
var cert = flag.String("cert", "", "Path to a certificate file.")
var pem = flag.String("key", "", "Path to a private key file.")
var en_https = flag.Bool("https", false, "Indicates whether TLS should be enabled.\nFor this option to take effect, a path to private key file and certificate file have to be provided.")
var en_http = flag.Bool("http", true, "Indicates whether HTTP should be enabled.")

func main() {
	flag.Parse()
	ws_srv := init_server(*cc)
	ws_srv.run(2)
	// go cli_ui(ws_srv)
	init_http(ws_srv, *ip, *port)
}
