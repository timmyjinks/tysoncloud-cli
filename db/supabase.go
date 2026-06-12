package db

import (
	"errors"
	"strings"

	"github.com/supabase-community/supabase-go"
)

func NewSupabaseStorage(url string, apiKey string) (*supabase.Client, error) {
	if strings.TrimSpace(url) == "" || strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("supabase url and api key are required")
	}
	cli, err := supabase.NewClient(url, apiKey, &supabase.ClientOptions{})
	if err != nil {
		return nil, err
	}

	return cli, nil
}
