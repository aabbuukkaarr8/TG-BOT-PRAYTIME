package store

import (
	_ "github.com/lib/pq"
	"github.com/wb-go/wbf/dbpg"
)

type Store struct {
	DB *dbpg.DB
}

func New() *Store {
	return &Store{}
}

func (s *Store) Open(config string) error {
	opts := dbpg.Options{
		MaxOpenConns: 1,
	}
	db, err := dbpg.New(config, nil, &opts)
	if err != nil {
		return err
	}

	if err := db.Master.Ping(); err != nil {
		return err
	}
	s.DB = db
	return nil
}

func (s *Store) Close() {
	err := s.DB.Master.Close()
	if err != nil {
		return
	}
}
