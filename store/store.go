package store

import (
	"database/sql"
	"github.com/supabase-community/supabase-go"
)

type SQLStoreService struct {
	db *sql.DB
}

type SupabaseStoreService struct {
	cli *supabase.Client
}

func NewSQLStoreService(db *sql.DB) *SQLStoreService {
	return &SQLStoreService{
		db: db,
	}
}

func NewSupabaseStoreService(cli *supabase.Client) *SupabaseStoreService {
	return &SupabaseStoreService{
		cli: cli,
	}
}

func (s *SQLStoreService) Close() {
	s.db.Close()
}
