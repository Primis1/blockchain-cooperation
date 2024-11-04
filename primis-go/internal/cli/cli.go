package cli

import (
	"blockchain/pkg/blockchain"
	"blockchain/pkg/utils"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
)
// NOTE `COMMAND` design pattern in use
type Command interface {
	Execute()
	Validate()
}

type BaseCommand struct {
	chain *blockchain.BlockChain
}


type CreateBlockchainCommand struct {
	address string
	chain   *blockchain.BlockChain
}
func (c *CreateBlockchainCommand) Execute(){
	chain := blockchain.
}


func (c *CreateBlockchainCommand) Validate(){
	if c.address == "" {
		utils.HandleErr("address is required")
	}
}




type SendCommand struct {
	from   string
	to     string
	amount int
	chain  *blockchain.BlockChain
}

type GetBalanceCommand struct {
	address string
	chain   *blockchain.BlockChain
}



type CommandLine struct {
	blockchain *blockchain.FacadeType
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" add -block BLOCK_DATA - add a block to the chain")
	fmt.Println(" print - Prints the blocks in the chain")

}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		// NOTE properly close connection to DB via goroutines
		runtime.Goexit()
	}
}

func (cli *CommandLine) addCliBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("Added Block!")

}

func (cli *CommandLine) printChain() {
	item := cli.blockchain.Iterator()

	for {
		block := item.Next()

		fmt.Printf("Prev. hash: %x\n", block.PreviousHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PreviousHash) == 0 {
			break
		}
	}

}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("add", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block data")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		utils.HandleErr(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if printChainCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}

		// NOTE we pass a string with pointer from user terminal
		cli.addCliBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}
