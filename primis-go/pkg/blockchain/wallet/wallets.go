package wallet

import "os"

type Wallets struct {
	Wallets map[string]*Wallet
}

var walletPath = os.Getenv("wallets")

func CreateWallets() *Wallet {
	wallets := Wallets{}

	wallets.Wallets = make(map[string]*Wallet)

	
}


func (w *Wallets) LoadFile() {
	
}