package main

import (
	"blockchain/internal/config"
	"blockchain/pkg/blockchain"
	"blockchain/pkg/logging"
	"fmt"
	"strconv"
)

// var chain = blockchain.Facade

func init() {
	// NOTE initialize env variables
	config.MustEnvironment()
	logging.GetLoggerInstance(logging.ERR)
	logging.GetLoggerInstance(logging.INFO)
}

func main() {

	chain := blockchain.Facade.InitBlockChain()

	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	for _, block := range chain.GetBlocks() {

		fmt.Printf("Previous Hash: %x\n", block.PreviousHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

	}
}
