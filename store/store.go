package store

import "database/sql"

type StoreService struct {
	db *sql.DB
}

func NewStoreService(db *sql.DB) *StoreService {
	return &StoreService{
		db: db,
	}
}

func (s *StoreService) Close() {
	s.db.Close()
}
