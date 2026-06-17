package store

import (
	"fmt"
	"strings"

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

func (s *SQLStoreService) UpdateServer(name string, newName string, description string, addr string) error {
	updateStr := []string{}
	args := []any{}
	argPos := 1

	if newName != "" {
		updateStr = append(updateStr, fmt.Sprintf("name = $%d", argPos))
		args = append(args, newName)
		argPos++
	}

	if description != "" {
		updateStr = append(updateStr, fmt.Sprintf("description = $%d", argPos))
		args = append(args, description)
		argPos++
	}

	if addr != "" {
		updateStr = append(updateStr, fmt.Sprintf("addr = $%d", argPos))
		args = append(args, addr)
		argPos++
	}

	if len(updateStr) == 0 {
		return nil
	}

	args = append(args, name)

	query := fmt.Sprintf("UPDATE servers SET %s WHERE name = $%d", strings.Join(updateStr, ", "), argPos)

	res, err := s.db.Exec(query, args...)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("server %q not found", name)
	}

	return nil
}

func (s *SQLStoreService) DeleteServer(names ...string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, name := range names {
		res, err := tx.Exec("DELETE FROM servers where name = $1", name)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return fmt.Errorf("server %q not found", name)
		}
	}

	return tx.Commit()
}
