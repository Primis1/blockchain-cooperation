package cli

import (
	"blockchain/pkg/blockchain"
	"blockchain/pkg/logging"
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

var info = logging.Info

type CreateBlockchainCommand struct {
	address string
	chain   *blockchain.BlockChain
}

func (c *CreateBlockchainCommand) Execute() {
	chain := blockchain.InitBlockChain(c.address)
	chain.Database.Close()
	info.Info("Finished creating blockchain!")
}

func (c *CreateBlockchainCommand) Validate() {
	if c.address == "" {
		utils.HandleErr("address is required")
	}
}

// for coin transfers
type SendCommand struct {
	BaseCommand
	from   string
	to     string
	amount int
}

func (c *SendCommand) Execute() {
	// 
}

type GetBalanceCommand struct {
	address string
	chain   *blockchain.BlockChain
}

type CommandLine struct {
	blockchain *blockchain.FacadeType
}


// TODO require refactor of blockchain package 
// TODO for further development   