package main

import (
	"log"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
	if app.store != nil {
		app.store.Close()
	}
}
