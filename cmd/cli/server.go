package main

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var name string
var description string
var addr string

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "add server resource",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(app)
		defer app.db.Close()

		if name == "" {
			return errors.New("server name empty")
		}

		if addr == "" {
			return errors.New("server addr empty")
		}

		uid, err := uuid.NewRandom()
		if err != nil {
			return err
		}

		if _, err := app.db.Exec("INSERT INTO servers (id, name, description, addr) VALUES ($1, $2, $3, $4)", uid, name, description, addr); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	serverCmd.Flags().StringVarP(&name, "name", "n", "", "name of server")
	serverCmd.Flags().StringVarP(&description, "description", "d", "", "description of server")
	serverCmd.Flags().StringVarP(&addr, "addr", "a", "", "ip address of server")
}
