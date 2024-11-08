package main

import (
	"blockchain/internal/cli"
	"blockchain/internal/config"
	"blockchain/pkg/logging"
)

// var chain = blockchain.Facade

func init() {
	// NOTE initialize env variables
	config.MustEnvironment()
	
	logging.Info.Info("log instance")
	logging.Err.Error("error instance")
}

func main() {
	cli := cli.NewCLI()
	cli.Run()

}
