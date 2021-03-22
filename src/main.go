package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	LastfmSharedSecret string
	LastfmApiKey       string
	Username           string
	RequestInterval    int
}

func cli_ui(srv *Server) {
	log.Printf("Console websocket client started\n")
	reader := bufio.NewReader(os.Stdin)
	for {
		in, _, _ := reader.ReadLine()
		fmt.Printf("Active goroutines: %d; connected clients: %d\n", runtime.NumGoroutine(), len(srv.clients))
		srv.broadcast <- string(in)
	}
}

var ip = flag.String("h", "0.0.0.0", "The address on which the server should listen. Possible values are: IPv4, IPv6, hostname")
var port = flag.Int("p", 8080, "The port of the server")
var cc = flag.Int("c", 1024, "Maximum number of clients connected concurrently to the server.")
var cert = flag.String("crt", "", "Path to a certificate file.")
var pem = flag.String("key", "", "Path to a private key file.")
var en_https = flag.Bool("https", false, "Indicates whether TLS should be enabled.\nFor this option to take effect, a path to private key file and certificate file have to be provided.")
var en_http = flag.Bool("http", true, "Indicates whether HTTP should be enabled.")
var en_cws = flag.Bool("console-ws-client", false, "Enables console writer to websocket clients for testing purposes.")
var cfg_f = flag.String("cfg", "./config.toml", "Path to a configuration file. For now, the only thing the file stores are last.fm api credentials.")

func load_config(cfg *Configuration) {
	_, err := os.Stat(*cfg_f)
	if os.IsNotExist(err) {
		os.Create(*cfg_f)
		f, err := os.OpenFile(*cfg_f, os.O_RDWR, 0600)
		if err != nil {
			log.Fatalf("Cannot open config file. Error: %s", err)
		}
		buf := new(bytes.Buffer)
		toml.NewEncoder(buf).Encode(&cfg)
		f.Write(buf.Bytes())
	}
	if _, err := toml.DecodeFile(*cfg_f, cfg); err != nil {
		log.Fatalf("An error occurred when reading the config file. Error: %s", err)
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
	init_http(ws_srv, *ip, *port)
}
