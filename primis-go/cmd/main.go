package main

import (
	"blockchain/internal/config"
	"blockchain/pkg/logging"
)

// var chain = blockchain.Facade

func init() {
	config.MustEnvironment()
	logging.GetLoggerInstance(logging.INFO)
	logging.GetLoggerInstance(logging.ERR)
}

func main() {}
