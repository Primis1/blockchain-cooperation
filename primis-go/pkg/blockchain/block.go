package blockchain

import (
	"blockchain/pkg/logging"
	"blockchain/pkg/sha"
	"blockchain/pkg/utils"
	"bytes"
	"encoding/gob"
)

var info = logging.Info
var errMsg = logging.Error

// NOTE each block contains huge number of transaction, to be created
type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	// NOTE field that indicates the "difficulty"
	Nonce int
}

// NOTE we hash each transaction of the block, transaction ID
func (b *Block) HashTransactions() []byte {
	var hashes [][]byte
	var hash [32]byte

	for _, tx := range b.Transactions {
		hashes = append(hashes, tx.ID)
	}

	hash = sha.ComputeHash(bytes.Join(hashes, []byte{}))
	return hash[:]	
}

// NOTE CreateBlock generates a new block with provided data and previous hash.
func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{
		Hash:         []byte{},
		Transactions: txs,
		PrevHash:     prevHash,
		Nonce:        0,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func CreateGenesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase},
		[]byte{})
}

// NOTE Principles of Serializing
// NOTE 1. create a buffer
// NOTE 2. create encoder 
// NOTE 3. encode 
// NOTE 4. send sequence of bytes  

func (b *Block) serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	utils.HandleErr(err)
	return res.Bytes()
}

// NOTE Principles of Deserializer 
// NOTE 1. declare the structure we want to make 
// NOTE 2. declare new decoder 
// NOTE 3. decode the structure 
// NOTE 4. return new structure 
func deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	utils.HandleErr(err)
	return &block
}
