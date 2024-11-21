package cli

import (
	"blockchain/pkg/blockchain"
	"blockchain/pkg/blockchain/wallet"
	"blockchain/pkg/network"
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
	fmt.Println(" send -from FROM -to TO -amount AMOUNT -mine - Send amount of coins. Then -mine flag is set, mine off of this node")
	fmt.Println(" createwallet - Creates a new Wallet")
	fmt.Println(" listaddresses - Lists the addresses in our wallet file")
	fmt.Println(" reindex - change the indexes of transactions")
	fmt.Println(" startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env. var. -miner enables mining")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) listAddresses(nodeId string) {
	wallets, _ := wallet.CreateWallets(nodeId)
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet(nodeId string) {
	wallets, _ := wallet.CreateWallets(nodeId)
	address := wallets.AddWallet()
	wallets.SaveFile(nodeId)

	fmt.Printf("New address is: %s\n", address)
}

func (cli *CommandLine) printChain(nodeId string) {
	chain := blockchain.ContinueBlockchain(nodeId)
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

func (cli *CommandLine) createBlockChain(address, nodeId string) {
	if !wallet.ValidateAddress("address") {
		utils.DisplayErr("Address is not valid")
	}
	chain := blockchain.InitBlockchain(address, nodeId)
	chain.Database.Close()

	UTXOSet := blockchain.UnspentTransactionSET{Blockchain: chain}
	UTXOSet.Reindex()

	fmt.Println("Finished!")

}

func (cli *CommandLine) getBalance(address, nodeId string) {
	if !wallet.ValidateAddress(address) {
		utils.DisplayErr("Address is not valid")
	}

	chain := blockchain.ContinueBlockchain(nodeId)
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

func (cli *CommandLine) StartNode(nodeID, minerAddress string) {
	fmt.Printf("Starting Node %s\n", nodeID)

	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			utils.DisplayErr("Wrong miner address!")
		}
	}
	network.StartServer(nodeID, minerAddress)
}

func (cli *CommandLine) send(from, to string, amount int, nodeId string, mineNow bool) {
	if !wallet.ValidateAddress(from) {
		utils.DisplayErr("Address is not valid")
	}
	if !wallet.ValidateAddress(to) {
		utils.DisplayErr("Address is not valid")
	}

	chain := blockchain.ContinueBlockchain(nodeId)
	// pass the reference to the blockchain
	UTXOSet := blockchain.UnspentTransactionSET{Blockchain: chain}
	defer chain.Database.Close()

	// create a transaction from followed arguments
	wallets, err := wallet.CreateWallets(nodeId)
	utils.DisplayErr(err)
	wallet := wallets.GetWallet(from)
	tx := blockchain.NewTransaction(wallet, to, amount, &UTXOSet)
	if mineNow {
		cbTx := blockchain.CoinbaseTx(from, "")
		txs := []*blockchain.Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else {
		network.SendTx(network.KnownNodes[0], tx)
		fmt.Println("send tx")
	}
	fmt.Println("Success!")
}

func (cli *CommandLine) reindexUTXO(nodeId string) {
	chain := blockchain.ContinueBlockchain(nodeId)
	defer chain.Database.Close()
	UTXOSet := blockchain.UnspentTransactionSET{Blockchain: chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountUnspentOuts()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		utils.DisplayErr("NODE_ID env is not set!")
		runtime.Goexit()
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	reindexCmd := flag.NewFlagSet("reindex", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	// further options
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward to ADDRESS")

	switch os.Args[1] {
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		utils.DisplayErr(err)
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		utils.DisplayErr(err)
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		utils.DisplayErr(err)
	case "reindex":
		err := reindexCmd.Parse(os.Args[2:])
		utils.DisplayErr(err)
	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		utils.DisplayErr(err)
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		utils.DisplayErr(err)
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		utils.DisplayErr(err)
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		utils.DisplayErr(err)
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress, nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}

	if createWalletCmd.Parsed() {
		cli.createWallet(nodeID)
	}
	if listAddressesCmd.Parsed() {
		cli.listAddresses(nodeID)
	}
	if reindexCmd.Parsed() {
		cli.reindexUTXO(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}

	if startNodeCmd.Parsed() {
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID, *startNodeMiner)
	}
}
