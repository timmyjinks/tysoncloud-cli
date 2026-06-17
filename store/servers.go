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

func (s *SQLStoreService) GetServers() ([]Server, error) {
	servers := make([]Server, 0)
	rows, err := s.db.Query("SELECT * FROM servers")
	if err != nil {
		return []Server{}, err
	}

	for rows.Next() {
		server := Server{}
		err := rows.Scan(&server.Id, &server.Name, &server.Description, &server.Addr)
		if err != nil {
			return []Server{}, err
		}
		servers = append(servers, server)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	return servers, nil
}

func (s *SQLStoreService) GetServerNames() ([]string, error) {
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
		if err := rows.Err(); err != nil {
			return []string{}, err
		}
		names = append(names, name)
	}

	return names, nil
}

func (s *SQLStoreService) GetServerByName(name string) (Server, error) {
	server := Server{}
	err := s.db.QueryRow("SELECT * FROM servers where name = $1", name).Scan(&server.Id, &server.Name, &server.Description, &server.Addr)
	if err != nil {
		return Server{}, err
	}
	return server, nil
}

func (s *SQLStoreService) AddServer(name string, description string, addr string) error {
	uid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	if _, err := s.db.Exec("INSERT INTO servers (id, name, description, addr) VALUES ($1, $2, $3, $4)", uid, name, description, addr); err != nil {
		return err
	}
	return nil
}

func (s *SQLStoreService) UpdateServer(name string, description string, addr string) error {
	_, err := s.db.Exec("UPDATE servers SET name = $1, description = $2, addr = $3", name, description, addr)
	if err != nil {
		return err
	}
	return nil
}

func (s *SQLStoreService) DeleteServer(names ...string) error {
	for _, name := range names {
		_, err := s.db.Exec("DELETE FROM servers where name = $1", name)
		if err != nil {
			return err
		}
	}
	return nil
}
