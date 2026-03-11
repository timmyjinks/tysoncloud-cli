package main

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/timmyjinks/tysoncloud-cli/internal/deploy"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [cmd]",
	Short: "short",
	Long:  "long",
	Args:  cobra.ArbitraryArgs,
	Run:   deployd,
}

func deployd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Usage()
		return
	}

	name := args[1]
	if name == "" {
		cmd.Usage()
		return
	}

	switch args[0] {
	case "create":
		err := deploy.Create(name)
		if err != nil {
			log.Println(err)
		}
	default:
		cmd.Usage()
	}
}

func init() {
}
