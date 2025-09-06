package main

import (
	"api/server"
	"flag"
)

var configPath = flag.String("config", "config.yaml", "path to config file")

func main() {
	flag.Parse()

	server.Serve(*configPath)
}
