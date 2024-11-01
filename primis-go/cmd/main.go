package main

import (
	"blockchain/internal/config"
	"blockchain/pkg/blockchain"
	"fmt"
)

var chain = blockchain.Facade

// var logs = logging.NewLogger(logging.INFO)
// var errs = logging.NewLogger(logging.ERR)

func init() {
	config.MustEnvironment()
}

func main() {

	chain.InitBlockChain()

	chain.AddBlock("idk")

	for _, block := range chain.Chain.Blocks {
		fmt.Printf("%s", block)
	}
}
