package main

import (
	"log"
)

func main() {
	if err := deployCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
