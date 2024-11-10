package wallet

import (
	"blockchain/pkg/utils"
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"os"
)

type Wallets struct {
	Wallets map[string]*Wallet
}

var (
	walletPath = os.Getenv("wallets")
	dbPath     = os.Getenv("dbPath")
)

func CreateWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.loadFile()

	return &wallets, err
}

// NOTE we generate new wallet with new private/pub keys
// NOTE then generate address
// NOTE assign new wallet to the hashmap, using address as a key
// NOTE it is within a Wallets structure
func (w *Wallets) AddWallet() string {
	wallet := makeWallet()
	address := string(wallet.Address())

	w.Wallets[address] = wallet

	return address
}

func (w Wallets) GetAllAddresses() []string {
	var arr = make([]string, 0, len(w.Wallets))

	for i := range w.Wallets {
		arr = append(arr, i)
	}

	return arr
}

func (w *Wallets) loadFile() error {
	if !utils.DirExist(dbPath) {
		errMsg.Error("wallet folder does not exist")
	}

	var wallets Wallets

	fileContent, err := os.ReadFile(walletPath)
	utils.HandleErr(err)

	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(fileContent))

	err = decoder.Decode(&wallets)
	utils.HandleErr(err)

	w.Wallets = wallets.Wallets

	return nil
}

func (w *Wallets) SaveFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)

	// We just send the the values of bytes into encoded buffer
	err := encoder.Encode(w)
	utils.HandleErr(err)

	// pass these bytes into a file
	err = os.WriteFile(walletPath, content.Bytes(), 0644)
	utils.HandleErr(err)
}
