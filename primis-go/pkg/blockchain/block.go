package blockchain

import (
	"blockchain/pkg/logging"
	"blockchain/pkg/sha"
	"blockchain/pkg/utils"
	"bytes"
	"encoding/gob"
)

var (
	info = logging.GetLoggerInstance(logging.INFO)
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	// NOTE field that indicates the "difficulty"
	Nonce int
}

type ProofOfWorkFactory struct{}

// NOTE struct for our abstract factory and factory object initialization
type BlockFactory struct {
	powFactory *ProofOfWorkFactory
}

// NOTE configuration for our factory

func newBlockFactory() *BlockFactory {
	return &BlockFactory{
		powFactory: &ProofOfWorkFactory{},
	}
}

type BlockConfig struct {
	Transaction []*Transaction
	PrevHash    []byte
	Diff        int
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha.ComputeHash(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// NOTE CreateBlock generates a new block with provided data and previous hash.
func (f *BlockFactory) CreateBlock(config BlockConfig) *Block {
	block := &Block{
		Hash:         []byte{},
		Transactions: config.Transaction,
		PrevHash:     config.PrevHash,
		Nonce:        0,
	}

	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (f *BlockFactory) CreateGenesis(coinbase *Transaction) *Block {
	return f.CreateBlock(BlockConfig{
		Transaction: []*Transaction{coinbase},
		PrevHash:    []byte{},
	})
}

func (b *Block) serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	utils.HandleErr(err)

	return res.Bytes()
}

func deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	utils.HandleErr(err)

	return &block
}
