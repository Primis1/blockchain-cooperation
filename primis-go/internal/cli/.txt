// DO NOT TOUCH MY CHEAT-SHEET!!!

package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"blockchain/pkg/blockchain"
)

// Command interface defines the contract for all blockchain commands
type Command interface {
	Execute() error
	Validate() error
}

// Base command structure for common fields
type BaseCommand struct {
	chain *blockchain.BlockChain
}

// CreateBlockchainCommand creates a new blockchain
type CreateBlockchainCommand struct {
	BaseCommand
	address string
}

func (c *CreateBlockchainCommand) Execute() error {
	chain := blockchain.InitBlockChain(c.address)
	chain.Database.Close()
	fmt.Println("Finished creating blockchain!")
	return nil
}

func (c *CreateBlockchainCommand) Validate() error {
	if c.address == "" {
		return fmt.Errorf("address is required")
	}
	return nil
}

// SendCommand handles coin transfers
type SendCommand struct {
	BaseCommand
	from   string
	to     string
	amount int
}

func (c *SendCommand) Execute() error {
	chain := blockchain.ContinueBlockChain(c.from)
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(c.from, c.to, c.amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Println("Success!")
	return nil
}

func (c *SendCommand) Validate() error {
	if c.from == "" || c.to == "" || c.amount <= 0 {
		return fmt.Errorf("invalid send parameters")
	}
	return nil
}

// GetBalanceCommand checks address balance
type GetBalanceCommand struct {
	BaseCommand
	address string
}

func (c *GetBalanceCommand) Execute() error {
	chain := blockchain.ContinueBlockChain(c.address)
	defer chain.Database.Close()

	balance := 0
	UTXOs := chain.FindUTXO(c.address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", c.address, balance)
	return nil
}

func (c *GetBalanceCommand) Validate() error {
	if c.address == "" {
		return fmt.Errorf("address is required")
	}
	return nil
}

// PrintChainCommand prints all blocks
type PrintChainCommand struct {
	BaseCommand
}

func (c *PrintChainCommand) Execute() error {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	
	iter := chain.Iterator()

	for {
		block := iter.Next()
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)

		for _, tx := range block.Transactions {
			fmt.Printf("\nTransaction %x:\n", tx.ID)
			for _, input := range tx.Inputs {
				fmt.Printf("  Input:\n")
				fmt.Printf("    TxID: %x\n", input.ID)
				fmt.Printf("    Out: %d\n", input.Out)
				fmt.Printf("    Signature: %s\n", input.Sig)
			}
			for _, output := range tx.Outputs {
				fmt.Printf("  Output:\n")
				fmt.Printf("    Value: %d\n", output.Value)
				fmt.Printf("    PubKey: %s\n", output.PubKey)
			}
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return nil
}

func (c *PrintChainCommand) Validate() error {
	return nil
}

// CommandExecutor handles command creation and execution
type CommandExecutor struct{}

func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{}
}

func (e *CommandExecutor) CreateCommand(args []string) (Command, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("no command provided")
	}

	switch args[1] {
	case "createblockchain":
		cmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
		address := cmd.String("address", "", "The address to send genesis block reward to")
		cmd.Parse(args[2:])
		return &CreateBlockchainCommand{address: *address}, nil

	case "send":
		cmd := flag.NewFlagSet("send", flag.ExitOnError)
		from := cmd.String("from", "", "Source wallet address")
		to := cmd.String("to", "", "Destination wallet address")
		amount := cmd.Int("amount", 0, "Amount to send")
		cmd.Parse(args[2:])
		return &SendCommand{from: *from, to: *to, amount: *amount}, nil

	case "getbalance":
		cmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
		address := cmd.String("address", "", "The address to get balance for")
		cmd.Parse(args[2:])
		return &GetBalanceCommand{address: *address}, nil

	case "printchain":
		cmd := flag.NewFlagSet("printchain", flag.ExitOnError)
		cmd.Parse(args[2:])
		return &PrintChainCommand{}, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", args[1])
	}
}

// CLI struct now uses CommandExecutor
type CLI struct {
	executor *CommandExecutor
}

func NewCLI() *CLI {
	return &CLI{
		executor: NewCommandExecutor(),
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" createblockchain -address ADDRESS - Create a blockchain and send genesis reward to address")
	fmt.Println(" getbalance -address ADDRESS - Get balance for address")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send amount from FROM address to TO")
	fmt.Println(" printchain - Prints all the blocks in the chain")
}

func (cli *CLI) Run() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}

	cmd, err := cli.executor.CreateCommand(os.Args)
	if err != nil {
		fmt.Println(err)
		cli.printUsage()
		runtime.Goexit()
	}

	if err := cmd.Validate(); err != nil {
		fmt.Println("Command validation failed:", err)
		runtime.Goexit()
	}

	if err := cmd.Execute(); err != nil {
		log.Panic("Command execution failed:", err)
	}
}