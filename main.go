package main

import (
	"blockchain/internal/cli"
	"blockchain/internal/config"
	"runtime"
	// "blockchain/internal/cli"
	// "runtime"
)

// var chain = blockchain.Facade

func init() {
	// NOTE initialize env variables
	config.MustEnvironment()
	config.CreateFilesIfNotExist()
}

func main() {
	defer runtime.Goexit()
	cli := cli.CommandLine{}
	cli.Run()
}
