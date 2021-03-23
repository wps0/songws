package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
)

func cli_ui(srv *Server) {
	log.Printf("Console websocket client started\n")
	reader := bufio.NewReader(os.Stdin)
	for {
		in, _, _ := reader.ReadLine()
		fmt.Printf("Active goroutines: %d; connected clients: %d\n", runtime.NumGoroutine(), len(srv.clients))
		srv.broadcast <- string(in)
	}
}

func main() {
	flag.Parse()
	cfg := Configuration{}
	load_config(&cfg)

	ws_srv := init_server(*cc)
	ws_srv.run(2)
	if *en_cws {
		go cli_ui(ws_srv)
	}
	go fetcher(ws_srv, &cfg)
	init_http(ws_srv, &cfg, *ip, *port)
}
