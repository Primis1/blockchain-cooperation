package cli

import (
	"blockchain/pkg/blockchain"
	"blockchain/pkg/blockchain/wallet"
	"blockchain/pkg/utils"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get the balance for an address")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain and sends genesis reward to address")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT - Send amount of coins")
	fmt.Println(" createwallet - Creates a new Wallet")
	fmt.Println(" listaddresses - Lists the addresses in our wallet file")
	fmt.Println(" reindex - change the indexes of transactions")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) listAddresses() {
	wallets, _ := wallet.CreateWallets()
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet() {
	wallets, _ := wallet.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()

	fmt.Printf("New address is: %s\n", address)
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Print(tx)
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockChain(address string) {
	if !wallet.ValidateAddress(address) {
		utils.HandleErr("Address is not valid")
	}
	chain := blockchain.InitBlockchain(address)
	chain.Database.Close()

	UTXOSet := blockchain.UnspentTransactionSET{Blockchain: chain}
	UTXOSet.Reindex()

	fmt.Println("Finished!")

}

func (cli *CommandLine) getBalance(address string) {
	if !wallet.ValidateAddress(address) {
		utils.HandleErr("Address is not valid")
	}

	chain := blockchain.ContinueBlockchain(address)
	UTXOSet := blockchain.UnspentTransactionSET{Blockchain: chain}
	defer chain.Database.Close()

	balance := 0
	pubKeyHash := utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	if !wallet.ValidateAddress(from) {
		utils.HandleErr("Address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		utils.HandleErr("Address is not valid")
	}

	chain := blockchain.ContinueBlockchain(from)
	// pass the reference to the blockchain
	UTXOSet := blockchain.UnspentTransactionSET{Blockchain: chain}
	defer chain.Database.Close()

	// create a transaction from followed arguments
	tx := blockchain.NewTransaction(from, to, amount, &UTXOSet)
	cbtx := blockchain.CoinbaseTx(from, "")
	block := chain.AddBlock([]*blockchain.Transaction{cbtx, tx})
	UTXOSet.Update(block)
	fmt.Println("Success!")
}

func (cli *CommandLine) reindexUTXO() {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	UTXOSet := blockchain.UnspentTransactionSET{Blockchain: chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountUnspentOuts()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexCmd := flag.NewFlagSet("reindex", flag.ExitOnError)

	// further options
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		utils.HandleErr(err)
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		utils.HandleErr(err)
	case "reindex":
		err := reindexCmd.Parse(os.Args[2:])
		utils.HandleErr(err)
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		utils.HandleErr(err)
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		utils.HandleErr(err)
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		utils.HandleErr(err)
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		utils.HandleErr(err)
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if reindexCmd.Parsed() {
		cli.reindexUTXO()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
