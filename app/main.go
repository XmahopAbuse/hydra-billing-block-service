package main

import (
	"flag"
	config2 "hydra-blocking/external/config"
	server "hydra-blocking/server"
	"log"
)

func main() {

	configFlag := flag.String("config", "config.yml", "Path to config file")
	flag.Parse()

	// Init config
	config, err := config2.NewConfig(*configFlag)
	if err != nil {
		log.Fatalln(err)
	}

	srv := server.NewServer(config)

	srv.RunServer()
}
