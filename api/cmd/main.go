package main

import (
	"api/internal/fx"
	"flag"
)

var configPath = flag.String("config", "config.yaml", "path to config file")

func main() {
	flag.Parse()

	app := fx.NewApp(*configPath)
	app.Run()
}
