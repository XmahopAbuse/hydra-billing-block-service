package store

import (
	"database/sql"
	"fmt"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
	"hydra-blocking/external/config"
	"log"
)

type Store struct {
	databaseURL string
	driver      string
	db          *sql.DB
	config      config.Config
	Repository  *LocalRepository
}

func NewStore(config *config.Config) *Store {
	databaseURL := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=disable password=%s user=%s", config.Store.Local.IPAddress,
		config.Store.Local.Port, config.Store.Local.Database, config.Store.Local.Password, config.Store.Local.Username)
	return &Store{databaseURL: databaseURL,
		driver: config.Store.Local.Driver}
}

func (s *Store) Open() error {
	var err error
	s.db, err = sql.Open(s.driver, s.databaseURL)
	if err != nil {
		return err
	}

	err = s.db.Ping()
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Migrate() error {
	if s.driver == "postgres" {
		log.Println("Running migrations...")
		driver, err := postgres.WithInstance(s.db, &postgres.Config{})
		if err != nil {
			return err
		}

		m, err := migrate.NewWithDatabaseInstance("file://migrations/", "postgres", driver)
		if err != nil {
			return err
		}

		if err = m.Up(); err != migrate.ErrNoChange && err != nil {
			return err
		}
		log.Println("Migrations apply")

	} else {
		fmt.Println("test")
	}

	return nil
}
