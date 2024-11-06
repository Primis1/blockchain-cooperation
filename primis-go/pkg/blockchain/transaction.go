package blockchain

import (
	"blockchain/pkg/sha"
	"blockchain/pkg/utils"
	"bytes"
	"encoding/gob"
	"fmt"
)

type Transaction struct {
	ID     []byte
	Inputs []TXI
	Output []TXO
}

type TXI struct {
	ID  []byte
	Out int
	Sig string
}

type TXO struct {
	Value int
	// allow user to share and receive coins
	Pubkey string
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txin := TXI{[]byte{}, -1, data}
	txout := TXO{100, to}

	tx := Transaction{nil, []TXI{txin}, []TXO{txout}}

	tx.Set()

	return &tx
}

// NOTE we create a hash based on the bytes of transaction
func (t *Transaction) Set() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(t)
	utils.HandleErr(err)
	hash = sha.ComputeHash(encoded.Bytes())

	// NOTE pass the values of array
	t.ID = hash[:]
}

func (tr *Transaction) IsCoinbase() bool {
	// NOTE if the transaction exist but is empty?
	return len(tr.Inputs) == 1 && len(tr.Inputs[0].ID) == 0 && tr.Inputs[0].Out == -1
}


// NOTE if returns true - account which holds
// NOTE data(`user_name`) owns the information 
func (in *TXI) CanUnlock(data string) bool {
	return in.Sig == data
}
func (out *TXO) CanBeUnlocked(data string) bool {
	return out.Pubkey == data
}
