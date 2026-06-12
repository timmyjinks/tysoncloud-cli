package store

import (
	"github.com/google/uuid"
)

type Server struct {
	Id          string
	Name        string
	Description string
	Addr        string
}

func (s *StoreService) GetServers() ([]Server, error) {
	servers := make([]Server, 0)
	rows, err := s.db.Query("SELECT * FROM servers")
	if err != nil {
		return []Server{}, err
	}
	defer rows.Close()

	for rows.Next() {
		server := Server{}
		err := rows.Scan(&server.Id, &server.Name, &server.Description, &server.Addr)
		if err != nil {
			return []Server{}, err
		}
		servers = append(servers, server)
	}

	return servers, nil
}

func (s *StoreService) GetServerNames() ([]string, error) {
	names := make([]string, 0)
	rows, err := s.db.Query("SELECT name FROM servers")
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	for rows.Next() {
		name := ""
		err := rows.Scan(&name)
		if err != nil {
			return []string{}, err
		}
		names = append(names, name)
	}

	return names, nil
}

func (s *StoreService) GetServerByName(name string) (Server, error) {
	server := Server{}
	err := s.db.QueryRow("SELECT * FROM servers where name = $1", name).Scan(&server.Id, &server.Name, &server.Description, &server.Addr)
	if err != nil {
		return Server{}, err
	}
	return server, nil
}

func (s *StoreService) AddServer(name string, description string, addr string) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	if _, err := s.db.Exec("INSERT INTO servers (id, name, description, addr) VALUES ($1, $2, $3, $4)", uid, name, description, addr); err != nil {
		return err
	}
	return nil
}

func (s *StoreService) UpdateServer(name string, description string, addr string) error {
	return nil
}

func (s *StoreService) DeleteServer(name string) error {
	return nil
}
