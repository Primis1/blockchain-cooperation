// NOTE The thing about blockchain and its transaction is, that it suppose to be a bunch of immutable objects
// NOTE which are constantly created and transfered, hence must be validated. When we already created an immutable piece of data
// NOTE we cant simply delete, undo or modify it, so its each move should be monitored and restricted.
// TODO Created a transaction -> Convert Into  BYTES -> Put into block -> Validate before put into block -> validate before sending transaction to someone -> So-on

package blockchain

import (
	"blockchain/pkg/blockchain/wallet"
	"blockchain/pkg/sha"
	"blockchain/pkg/utils"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
)

type Transaction struct {
	ID     []byte
	Inputs []TXI
	Output []TXO
}

type TXI struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

type TXO struct {
	Value int
	// allow user to share and receive coins
	PubkeyHash []byte
}

func (in *TXI) UserKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKey(pubKeyHash)

	return bytes.Equal(lockingHash, pubKeyHash)
}

func (out *TXO) Lock(address []byte) {
	pubKeyHash := utils.Base58Decode(address)

	// remove version byte and last four
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

	out.PubkeyHash = pubKeyHash
}

// NOTE we unlock the block if the pubKey of a user is the same
// NOTE as pubKey which is inside the transaction
func (out *TXO) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubkeyHash, pubKeyHash)
}

// NOTE we receive a string from command line, so have to convert it
// NOTE + we compare the "rights" on the transaction,
// NOTE if both: owner-hash and transaction which was in output

func NewTXO(value int, address string) *TXO {
	txo := &TXO{value, nil}

	txo.IsLockedWithKey([]byte(address))
	return txo
}

// NOTE Convert transaction into slice of bytes
// NOTE Just like with blocks
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)

	utils.HandleErr(err)

	return encoded.Bytes()
}

// Yet again we hash transaction
func (t *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *t
	txCopy.ID = []byte{}

	hash = sha.ComputeHash(txCopy.Serialize())

	return hash[:]

}

// NOTE sign and verify transactions. Similar to wallets, dunno yet what our key represents
func (t *Transaction) Sign(private ecdsa.PrivateKey, prevT map[string]Transaction) {
	if t.IsCoinbase() {
		return
	}

	for _, in := range t.Inputs {

		// NOTE to access the outputs we should iterate through inputs
		// NOTE bcause they contain reference to the OUTs. Therefore if some input
		// NOTE contains no values, it means that it does not exist
		if prevT[hex.EncodeToString(in.ID)].ID == nil {
			utils.HandleErr("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := t.TrimmedCopy()

	// NOTE idea is that we every transactions besides current are empty
	for inId, in := range txCopy.Inputs {
		//NOTE convert bytes into string/key
		prevTx := prevT[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Output[in.Out].PubkeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &private, txCopy.ID)
		utils.HandleErr(err)

		// NOTE we `sign` ID, which is hash
		signature := append(r.Bytes(), s.Bytes()...)

		t.Inputs[inId].Signature = signature
	}
}

func (t *Transaction) TrimmedCopy() Transaction {
	var input []TXI
	var output []TXO

	for _, in := range t.Inputs {
		// NOTE  							clear out keys
		input = append(input, TXI{in.ID, in.Out, nil, nil})
	}

	for _, out := range t.Output {
		output = append(output, TXO{out.Value, out.PubkeyHash})
	}
	return Transaction{t.ID, input, output}
}

func (t *Transaction) Verify(prevT map[string]Transaction) bool {
	if t.IsCoinbase() {
		return true
	}

	for _, in := range t.Inputs {
		if prevT[hex.EncodeToString(in.ID)].ID == nil {
			errMsg.Error("Previous transaction does not exist")
		}
	}

	txCopy := t.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range t.Inputs {
		prevTx := prevT[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Output[in.Out].PubkeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[inId].PubKey = nil

		r := big.Int{}
		s := big.Int{}

		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}

func NewTransaction(from, to string, amount int, chain *BlockChain) *Transaction {
	var inputs []TXI
	var outputs []TXO

	wallets, err := wallet.CreateWallets()
	utils.HandleErr(err)

	w := wallets.GetWallet(from)

	pubKey := wallet.PublicKey(w.PublicKey)

	acc, validity := chain.FindSpendableOutputs(pubKey, amount)

	if acc < amount {
		utils.HandleErr("Error: not enough funds")
	}

	for txid, outs := range validity {
		txID, err := hex.DecodeString(txid)
		utils.HandleErr(err)

		for _, out := range outs {
			input := TXI{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTXO(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTXO(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()

	// NOTE sign transaction with our Prv Key
	chain.SignTransaction(&tx, w.PrivateKey)

	return &tx
}

func (tr *Transaction) IsCoinbase() bool {
	// NOTE if the transaction exist but is empty?
	return len(tr.Inputs) == 1 && len(tr.Inputs[0].ID) == 0 && tr.Inputs[0].Out == -1
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

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txin := TXI{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXO(100, to)

	tx := Transaction{nil, []TXI{txin}, []TXO{*txout}}
	tx.Set()

	return &tx
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TransactionID: %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       	 %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature:	 %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    	 %x", input.PubKey))
	}

	for i, output := range tx.Output {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubkeyHash))
	}

	return strings.Join(lines, "\n")
}
