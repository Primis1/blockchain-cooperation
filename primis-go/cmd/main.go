package main

import (
	"blockchain/internal/cli"
	"blockchain/internal/config"
	"runtime"
)

// var chain = blockchain.Facade

func init() {
	// NOTE initialize env variables
	config.MustEnvironment()
}

func main() {
	defer runtime.Goexit()
	cli := cli.CommandLine{}
	cli.Run()
}
