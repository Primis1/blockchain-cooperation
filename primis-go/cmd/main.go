package main

import (
	"blockchain/internal/cli"
	"blockchain/internal/config"
)

// var chain = blockchain.Facade

func init() {
	// NOTE initialize env variables
	config.MustEnvironment()
}

func main() {
	cli := cli.NewCLI()

	cli.Run()
}
