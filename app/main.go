package main

import (
	config2 "hydra-blocking/external/config"
	server "hydra-blocking/server"
	"log"
)

func main() {

	// Init config
	config, err := config2.NewConfig("config.yml")
	if err != nil {
		log.Fatalln(err)
	}

	srv := server.NewServer(config)

	srv.RunServer()
}
