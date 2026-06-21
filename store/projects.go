package store

import (
	"encoding/json"
	"time"

	"github.com/supabase-community/postgrest-go"
)

type ProjectsTable struct {
	ID           string    `json:"id,omitempty"`
	DeploymentId string    `json:"deployment_id,omitempty"`
	Namespace    string    `json:"namespace,omitempty"`
	Name         string    `json:"name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

func (s *SupabaseStoreService) GetProjects() ([]ProjectsTable, error) {
	res, _, err := s.cli.From("projects").Select("*", "exact", false).Order("created_at", &postgrest.OrderOpts{Ascending: false}).Execute()
	if err != nil {
		return nil, err
	}

	var table []ProjectsTable = []ProjectsTable{}
	if err := json.Unmarshal(res, &table); err != nil {
		return nil, err
	}

	return table, nil
}
