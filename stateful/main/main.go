package main

import (
	"fmt"
	"gitlab.com/eper.io/engine/metadata"
	"gitlab.com/eper.io/engine/stateful"
	"net/http"
	"os"
)

// This document is Licensed under Creative Commons CC0.
// To the extent possible under law, the author(s) have dedicated all copyright and related and neighboring rights
// to this document to the public domain worldwide.
// This document is distributed without any warranty.
// You should have received a copy of the CC0 Public Domain Dedication along with this document.
// If not, see https://creativecommons.org/publicdomain/zero/1.0/legalcode.

// Usage: go run main.go [:port]
// Example: go run main.go :8080
func main() {
	port := metadata.Http11Port
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	key, ok := os.LookupEnv("KEY")
	if ok && key != "" {
		metadata.ActivationKey = key
	}

	stateful.SetupPersistence()

	err := http.ListenAndServe(port, nil)
	if err != nil {
		fmt.Println(err)
		fmt.Println("usage: go run stateful/main/main.go :7777")
	}
}
