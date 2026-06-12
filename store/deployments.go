package store

import (
	"encoding/json"
	"time"

	"github.com/supabase-community/postgrest-go"
)

type DeploymentTable struct {
	ID          string    `json:"id,omitempty"`
	ContainerID string    `json:"container_id,omitempty"`
	UserID      string    `json:"user_id,omitempty"`
	Name        string    `json:"name,omitempty"`
	URL         string    `json:"url,omitempty"`
	Source      string    `json:"source,omitempty"`
	Status      string    `json:"status,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Type        string    `json:"type,omitempty"`
	Volume      string    `json:"volume,omitempty"`
}

func (s *SupabaseStoreService) GetDeployments() ([]DeploymentTable, error) {
	res, _, err := s.cli.From("deployments").Select("*", "exact", false).Order("created_at", &postgrest.OrderOpts{Ascending: false}).Execute()
	if err != nil {
		return nil, err
	}

	var table []DeploymentTable = []DeploymentTable{}
	if err := json.Unmarshal(res, &table); err != nil {
		return nil, err
	}

	return table, nil
}
