package main

import (
	"blockchain/internal/config"
	"blockchain/internal/singleton"
	"blockchain/pkg/blockchain"
)

var chain = blockchain.Facade

func init() {
	config.MustEnvironment()
}


func main() {

	chain.InitBlockChain()

	chain.AddBlock("idk")

	for _, block := range chain.Chain.Blocks {
		singleton.Log("%s", block)
	}
}
