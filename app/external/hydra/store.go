package hydra

import (
	"database/sql"
	"fmt"
	_ "github.com/godror/godror"
	"hydra-blocking/external/config"
)

type Store struct {
	db         *sql.DB
	config     *config.Config
	Repository *HydraRepository
}

func NewStore(conf *config.Config) (*Store, error) {
	var store Store
	var err error
	datebaseURL := fmt.Sprintf(`user="%s" password="%s" connectString="%s/%s"`,
		conf.Store.Hydra.Username, conf.Store.Hydra.Password, conf.Store.Hydra.IPAddress, conf.Store.Hydra.Service)

	store.db, err = sql.Open("godror", datebaseURL)
	if err != nil {
		return nil, err
	}

	err = store.db.Ping()
	if err != nil {
		return nil, err
	}

	store.config = conf

	store.Repository = InitRepository(&store)

	return &store, nil
}

func (s *Store) Close() {
	s.db.Close()
}
