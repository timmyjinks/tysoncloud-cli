package store

import (
	"encoding/json"
	"time"

	"github.com/supabase-community/postgrest-go"
)

type DeploymentTable struct {
	ID        string    `json:"id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Name      string    `json:"name,omitempty"`
	URL       string    `json:"url,omitempty"`
	Status    string    `json:"status,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *SupabaseStoreService) GetDeployments() ([]DeploymentTable, error) {
	res, _, err := s.cli.From("deployments_new").Select("*", "exact", false).Order("created_at", &postgrest.OrderOpts{Ascending: false}).Execute()
	if err != nil {
		return nil, err
	}

	var table []DeploymentTable = []DeploymentTable{}
	if err := json.Unmarshal(res, &table); err != nil {
		return nil, err
	}

	return table, nil
}
