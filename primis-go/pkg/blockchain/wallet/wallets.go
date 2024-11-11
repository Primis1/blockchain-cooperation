// BUG type elliptic.p256Curve has no exported fields
// NOTE For data to encoded and decoded, field should be `exportable`.
// NOTE `gob` package can not access the field which are from lower case
package wallet

import (
	"blockchain/pkg/utils"
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"encoding/json"
	"os"
)

type Wallets struct {
	Wallets map[string]*Wallet
}

// TODO remove boiler-plate 
var (
	walletPath = "../tmp/wallet.data"
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

func (w *Wallets) GetWallet(address string) *Wallet {
	return w.Wallets[address]
}

func (w Wallets) GetAllAddresses() []string {
	var arr = make([]string, 0, len(w.Wallets))

	for i := range w.Wallets {
		arr = append(arr, i)
	}

	return arr
}

func (w *Wallets) loadFile() error {
	if !utils.DirExist(walletPath) {
		file, _ := os.Open(walletPath)

		defer file.Close()
	}

	var wallets Wallets

	fileContent, err := os.ReadFile(walletPath)
	utils.HandleErr(err)

	gob.Register(elliptic.P256())

	decoder := gob.NewDecoder(bytes.NewReader(fileContent))

	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}

	w.Wallets = wallets.Wallets

	return nil
}

func (w *Wallets) SaveFile() {
	jsonData, err := json.Marshal(w)
	utils.HandleErr(err)

	err = os.WriteFile(walletPath, jsonData, 0544)
	utils.HandleErr(err)
}
