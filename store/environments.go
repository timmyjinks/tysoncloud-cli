package store

import (
	"encoding/json"
	"time"
)

type EnvironmentsTable struct {
	ID        string    `json:"id,omitempty"`
	ServiceId string    `json:"service_id,omitempty"`
	Key       string    `json:"key,omitempty"`
	Val       string    `json:"val,omitempty"`
	Secret_Id string    `json:"secret_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *SupabaseStoreService) GetEnvironments(id string) ([]EnvironmentsTable, error) {
	res := s.cli.Rpc("get_environments", "", map[string]string{
		"p_service_id": id,
	})

	var table []EnvironmentsTable = []EnvironmentsTable{}
	if err := json.Unmarshal([]byte(res), &table); err != nil {
		return nil, err
	}

	return table, nil
}

