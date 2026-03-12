package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/timmyjinks/tysoncloud-cli/internal/deploy"
)

var KEY = "$HOME/config/tsoncloud-cli/credentials.json"

var deployCmd = &cobra.Command{
	Use:   "tysoncloud-cli [command]",
	Short: "short",
	Long:  "long",
	Args:  cobra.ArbitraryArgs,
	Run:   deployRun,
}

var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "short create",
	Long:  "create long",
	Args:  cobra.ArbitraryArgs,
	Run:   createRun,
}

var updateCmd = &cobra.Command{
	Use:   "update [name] [new name]",
	Short: "short update",
	Long:  "upate long",
	Args:  cobra.ArbitraryArgs,
	Run:   updateRun,
}

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "short delete",
	Long:  "delete long",
	Args:  cobra.ArbitraryArgs,
	Run:   deleteRun,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login",
	Long:  "login",
	Args:  cobra.ArbitraryArgs,
	Run:   loginRun,
}

var logout = &cobra.Command{
	Use:   "logout",
	Short: "logout",
	Long:  "logout",
	Args:  cobra.ArbitraryArgs,
	Run:   createRun,
}

func deployRun(cmd *cobra.Command, args []string) {
}

func createRun(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	name := args[0]
	if name == "" {
		return
	}

	err := deploy.Create(name)
	if err != nil {
		log.Println(err)
		return
	}
}

func updateRun(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		cmd.Usage()
		return
	}

	name := args[0]
	if name == "" {
		cmd.Usage()
		return
	}

	newName := args[1]
	if newName == "" {
		cmd.Usage()
		return
	}

	err := deploy.Update(name, newName)
	if err != nil {
		log.Println(err)
		return
	}
}

func deleteRun(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		return
	}

	name := args[0]
	if name == "" {
		cmd.Usage()
		return
	}

	err := deploy.Delete(name)
	if err != nil {
		log.Println(err)
		return
	}
}

func loginRun(cmd *cobra.Command, args []string) {
	fmt.Println(args[0])
	fmt.Println(args[1])

	reqBody := struct {
		Name     string
		Password string
	}{
		Name:     args[0],
		Password: args[1],
	}

	b, err := json.Marshal(reqBody)
	if err != nil {
		log.Println(err)
		return
	}

	resp, err := http.Post("http://tysoncloud-rewrite-test.tysonjenkins.dev/auth/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println(err)
		return
	}

	bb, _ := io.ReadAll(resp.Body)
	fmt.Println(resp)
	fmt.Println(string(bb))
}

func init() {
	deployCmd.AddCommand(createCmd)
	deployCmd.AddCommand(updateCmd)
	deployCmd.AddCommand(deleteCmd)
	deployCmd.AddCommand(loginCmd)
}
