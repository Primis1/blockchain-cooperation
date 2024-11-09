package main

import (
	"blockchain/internal/config"
	"blockchain/pkg/blockchain/wallet"
	"blockchain/pkg/logging"
)

// var chain = blockchain.Facade

func init() {
	// NOTE initialize env variables
	config.MustEnvironment()

	logging.Info.Info("log instance")
}

func main() {
	cli := wallet.MakeWallet()
	cli.Address()
}
