package main

import (
	"blockchain/internal/config"
	"blockchain/pkg/logging"
)

// var chain = blockchain.Facade

func init() {
	// NOTE initialize env variables
	config.MustEnvironment()
	logging.GetLoggerInstance(logging.ERR)
	logging.GetLoggerInstance(logging.INFO)
}

func main() {}
