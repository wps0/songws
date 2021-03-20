package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	LastfmSharedSecret string
	LastfmApiKey       string
	Username           string
	RequestInterval    int
}

func cli_ui(srv *Server) {
	fmt.Printf("Console websocket client started\n")
	for {
		var in string
		fmt.Scan(&in)
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
var en_cws = flag.Bool("console-ws-client", false, "Enables console websocket client for testing purposes.")
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
