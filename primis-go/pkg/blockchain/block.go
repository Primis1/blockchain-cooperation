package blockchain

import (
	"blockchain/pkg/sha"
	"blockchain/pkg/utils"
	"bytes"
	"encoding/gob"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

type ProofOfWorkFactory struct{}

type BlockFactory struct {
	powFactory *ProofOfWorkFactory
}

func NewBlockFactory() *BlockFactory {
	return &BlockFactory{
		powFactory: &ProofOfWorkFactory{},
	}
}

type BlockConfig struct {
	Transaction []*Transaction
	PrevHash    []byte
	Diff        int
}

// NOTE GenesisBlock creates the initial block in the blockchain.
// NOTE We should declare serializer for default GO DB
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

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	utils.HandleErr(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	utils.HandleErr(err)

	return &block
}


