package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type Server struct {
	Id          string
	Name        string
	Description string
	Addr        string
}

var description string

var addServerCmd = &cobra.Command{
	Use:   "server [name] [address]",
	Short: "add server resource",
	Args:  cobra.ExactArgs(2),
	Annotations: map[string]string{
		"database": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		addr := args[1]
		if err := app.store.AddServer(name, description, addr); err != nil {
			return err
		}
		return nil
	},
}

var getServerCmd = &cobra.Command{
	Use:   "servers [name]",
	Short: "get server resources",
	Args:  cobra.MaximumNArgs(1),
	Annotations: map[string]string{
		"database": "true",
	},
	ValidArgsFunction: getServerComplete,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			name := args[0]
			server, err := app.store.GetServerByName(name)
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tDESCRIPTION\tADDRESS")
			fmt.Fprintf(w, "%v\t%v\t%v\n", server.Name, server.Description, server.Addr)
			w.Flush()

			return nil
		}

		if len(args) == 0 {
			servers, err := app.store.GetServers()
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "NAME\tDESCRIPTION\tADDRESS")
			for _, server := range servers {
				fmt.Fprintf(w, "%v\t%v\t%v\n", server.Name, server.Description, server.Addr)
			}
			w.Flush()
		}
		return nil
	},
}

func init() {
	addServerCmd.Flags().StringVarP(&description, "description", "d", "", "description of server")
}
