package server

import (
	"github.com/gorilla/mux"
	"hydra-blocking/external/config"
	"hydra-blocking/external/hydra"
	"hydra-blocking/store"
	"log"
	"net/http"
)

type Server struct {
	config     *config.Config
	hydraStore *hydra.Store
	router     *mux.Router
	localStore *store.Store
}

func NewServer(config *config.Config) *Server {
	return &Server{config: config}
}

func (s *Server) RunServer() {
	var err error

	// Init hydra store
	s.hydraStore, err = hydra.NewStore(s.config)
	defer s.hydraStore.Close()

	if err != nil {
		log.Fatalln(err)
	}

	// Init local store
	s.localStore = store.NewStore(s.config)
	err = s.localStore.Open()
	if err != nil {
		log.Fatalln(err)
	}
	// Init local repository
	s.localStore.Repository = store.InitLocalRepository(s.localStore)

	defer s.localStore.Close()

	// Run local migrations
	err = s.localStore.Migrate()
	if err != nil {
		log.Fatalln(err)
	}

	// Init router
	s.initRouter()

	// Run http server
	err = http.ListenAndServe(":8080", s.router)
	if err != nil {
		log.Fatalln(err)
	}
}

func (s *Server) initRouter() {
	s.router = mux.NewRouter()

	s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	s.router.HandleFunc("/test", s.MainHandler)
	s.router.HandleFunc("/setBlock", s.SetBlockHandler).Methods("POST")
	s.router.HandleFunc("/removeBlock", s.RemoveBlockHandler).Methods("POST")
	s.router.HandleFunc("/getStatus", s.GetStatusHandler).Methods("POST")

}
