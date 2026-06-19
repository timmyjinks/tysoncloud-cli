package main

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var pingCmd = &cobra.Command{
	Use:               "ping",
	Short:             "ping server",
	Args:              cobra.ExactArgs(1),
	ValidArgsFunction: getServerComplete,
	Annotations: map[string]string{
		"database": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		server, err := app.store.GetServerByName(name)
		if err != nil {
			return err
		}

		command := exec.Command("ping", "-c", "1", server.Addr)
		if err := command.Run(); err != nil {
			return err
		}
		fmt.Printf("Server %s pinged successfully\n", server.Addr)
		return nil
	},
}
