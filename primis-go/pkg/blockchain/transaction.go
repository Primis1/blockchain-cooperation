package blockchain

import (
	"blockchain/pkg/logging"
	"blockchain/pkg/sha"
	"blockchain/pkg/utils"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"runtime"
)

var info = logging.Info

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

func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TXI
	var outputs []TXO

	acc, valid := chain.FindSpendableOutputs(from, amount)

	if acc < amount {
		info.Info("Not enough funds")
		runtime.Goexit()
	}

	for txid, outs := range valid {
		txID, err := hex.DecodeString(txid)
		utils.HandleErr(err)

		for _, out := range outs {
			input := TXI{txID, out, from}
			inputs = append(inputs, input)

		}
	}

	outputs = append(outputs, TXO{amount, to})

	if acc > amount {
		outputs = append(outputs, TXO{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}

	tx.Set()

	return &tx
}

// NOTE if returns true - account which holds
// NOTE data(`user_name`) owns the information
func (in *TXI) CanUnlock(data string) bool {
	return in.Sig == data
}
func (out *TXO) CanBeUnlocked(data string) bool {
	return out.Pubkey == data
}
