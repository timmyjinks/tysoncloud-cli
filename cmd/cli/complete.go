package main

import (
	"github.com/spf13/cobra"
)

func getServerComplete(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if app.store == nil {
		return nil, cobra.ShellCompDirectiveError
	}
	names, err := app.store.GetServerNames()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return names, cobra.ShellCompDirectiveNoFileComp
}
