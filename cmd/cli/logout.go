package main

import "github.com/spf13/cobra"

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "logout of tysoncloud account",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
