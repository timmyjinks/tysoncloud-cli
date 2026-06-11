package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "login of tysoncloud account",
	Args:  cobra.ArbitraryArgs,
	Run:   loginRun,
}

func loginRun(cmd *cobra.Command, args []string) {
	defer app.db.Close()
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

	resp, err := http.Post("http://tysoncloud.tysonjenkins.dev/auth/login", "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println(err)
		return
	}

	bb, _ := io.ReadAll(resp.Body)
	fmt.Println(resp)
	fmt.Println(string(bb))
}
