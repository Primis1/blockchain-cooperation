package cli

import (
	"blockchain/pkg/blockchain"
	"blockchain/pkg/logging"
	"blockchain/pkg/utils"
	"flag"
	"fmt"
	"runtime"
)

var errMsg = logging.Err
var info = logging.Info

type Command interface {
	Execute() error
	Validate() error
}

type CLIFacade struct {
	repo    blockchain.BlockchainRepository
	service *blockchain.BlockchainService
	address string
}

type BaseCommand struct {
	facade *CLIFacade
}

// CreateBlockchainCommand creates a new blockchain
type CreateBlockchainCommand struct {
	BaseCommand
	address string
}

func NewFacadeFactory() *CLIFacade {
	repo := blockchain.NewBlockchainRepository()
	service := blockchain.NewBlockchainService(repo)

	return &CLIFacade{
		service: service,
	}
}

func (c *CreateBlockchainCommand) Execute() error {
	coinbase := blockchain.CoinbaseTx(c.address, "")
	block := c.facade.service.Factory.CreateGenesis(coinbase)

	err := c.facade.repo.SaveBlock(block)
	errMsg.Error("\nfailed to save genesis block \n", err)

	err = c.facade.repo.SaveLastHash(block.Hash)
	errMsg.Error("\nfailed to save last hash\n", err)

	info.Info("Finished creating blockchain")
	return nil
}

func (c *CreateBlockchainCommand) Validate() error {
	if c.address == "" {
		utils.HandleErr("address is required")
	}

	return nil
}

type SendCommand struct {
	BaseCommand
	from   string
	to     string
	amount int
}

func (c *SendCommand) Execute() error {
	accumulated, validOutputs := c.facade.repo.FindSpendableOutputs(c.from, c.amount)

	if accumulated < c.amount {
		errMsg.Error("insufficient funds: got %d, need %d", accumulated, validOutputs)
	}

	// transactions
	var inputs []blockchain.TXI
	var outputs []blockchain.TXO

	// NOTE reminder -  FindSpendableOutputs() returns hashmap
	for txid, outs := range validOutputs {
		for _, out := range outs {

			// NOTE fill up the new slice with hash-keys and int-slices
			input := blockchain.TXI{
				ID:  []byte(txid),
				Out: out,
				Sig: c.from,
			}

			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, blockchain.TXO{Value: c.amount, Pubkey: c.to})

	if accumulated > c.amount {
		outputs = append(outputs, blockchain.TXO{Value: accumulated - c.amount, Pubkey: c.from})
	}

	// NOTE filled an object with data
	tx := &blockchain.Transaction{
		ID:     nil,
		Inputs: inputs,
		Output: outputs,
	}

	// NOTE add the transaction into blockchain
	err := c.facade.service.AddBlock([]*blockchain.Transaction{tx})
	errMsg.Error("failed to add block: %w", err)

	info.Info("Transfer been successful")

	// we do return nil, to finish the program
	return nil
}

func (c *SendCommand) Validate() error {
	if c.from == "" || c.to == "" || c.amount <= 0 {
		utils.HandleErr("invalid send parameters")
		return nil
	}
	return nil
}

type GetBalanceCommand struct {
	BaseCommand
	address string
}

func (c *GetBalanceCommand) Execute() error {
	// NOTE collect data about user transactions
	uniqueUnspentTransaction := c.facade.repo.FindUnspentTransactionsOutputs(c.address)

	balance := 0

	// NOTE simply sum "money" of all transaction together
	for _, out := range uniqueUnspentTransaction {
		balance += out.Value
	}
	info.Info("Balance of %s: %d\n", c.address, balance)
	return nil
}

func (c *GetBalanceCommand) Validate() error {
	if c.address == "" {
		errMsg.Error("Address is required")
	}
	return nil
}

func (f *CommandFactory) CreateCommand(args []string) (Command, error) {
	if len(args) < 2 {
		errMsg.Error("no command provided")
		return nil, nil
	}

	switch args[1] {
	case "cb":
		cmd := flag.NewFlagSet("cs", flag.ExitOnError)

		address := cmd.String("address", "", "the address to send genesis block reward to")

		cmd.Parse(args[2:])

		return &CreateBlockchainCommand{
			BaseCommand: BaseCommand{facade: f.facade},
			address:     *address,
		}, nil

	case "send":
		cmd := flag.NewFlagSet("send", flag.ExitOnError)

		from := cmd.String("from", "", "Source wallet address")
		to := cmd.String("to", "", "Destination wallet address")

		amount := cmd.Int("amount", 0, "Amount to send")
		cmd.Parse(args[2:])

		return &SendCommand{
			BaseCommand: BaseCommand{facade: f.facade},
			from:        *from,
			to:          *to,
			amount:      *amount,
		}, nil
	case "getbalance":
		cmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

		address := cmd.String("address", "", "The address to get")
		cmd.Parse(args[2:])

		return &GetBalanceCommand{
			BaseCommand: BaseCommand{facade: f.facade},
			address:     *address,
		}, nil
	default:
		errMsg.Error("unknown command")
		return nil, nil
	}
}

type CLI struct {
	factory *CommandFactory
}

type CommandFactory struct {
	facade *CLIFacade
}

// NOTE we build a factory from the facade
func NewCommandFactory(facade *CLIFacade) *CommandFactory {
	return &CommandFactory{facade: facade}
}

func NewBlockchainFacade() *CLIFacade {
	repo := blockchain.NewBlockchainRepository()
	service := blockchain.NewBlockchainService(repo)

	return &CLIFacade{
		service: service,
		repo:    repo,
	}
}

func NewCLI() *CLI {
	facade := NewBlockchainFacade()

	return &CLI{
		factory: NewCommandFactory(facade),
	}
}

func (cli CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" createblockchain -address ADDRESS - Create a blockchain and send genesis reward to address")
	fmt.Println(" getbalance -address ADDRESS - Get balance for address")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send amount from FROM address to TO")

}

func (cli *CLI) Run() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}

	cmd, err := cli.factory.CreateCommand(os.Args)
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
