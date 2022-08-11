package main

import (
	"fmt"
	"os"
)

//go:generate go run ./

func main() {
	if err := generateSDK(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
